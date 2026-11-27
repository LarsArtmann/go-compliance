package discovery_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/discovery"
)

// unixTestServer serves HTTP over a unix socket the way the daemon does.
type unixTestServer struct {
	listener net.Listener
	server   *http.Server
	socket   string
}

func newUnixTestServer(handler http.Handler) *unixTestServer {
	dir, err := os.MkdirTemp("", "compliance-discovery-*")
	Expect(err).NotTo(HaveOccurred())

	socket := filepath.Join(dir, "daemon.sock")
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socket)
	Expect(err).NotTo(HaveOccurred())

	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}

	go func() { _ = server.Serve(listener) }()

	return &unixTestServer{listener: listener, server: server, socket: socket}
}

func (s *unixTestServer) close() {
	_ = s.server.Close()
	_ = os.RemoveAll(filepath.Dir(s.socket))
}

var _ = Describe("discovery client", func() {
	var srv *unixTestServer

	BeforeEach(func() {
		mux := http.NewServeMux()

		mux.HandleFunc("POST /v1/discover", func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

			var request struct {
				SearchPaths []string `json:"searchPaths"`
			}

			requestBody, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Unmarshal(requestBody, &request)).To(Succeed())
			Expect(request.SearchPaths).To(Equal([]string{"/home/lars/projects"}))

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"projects":[
				{"dirName":"go-finding","path":"/home/lars/projects/go-finding"},
				{"dirName":"go-compliance","path":"/home/lars/projects/go-compliance"}
			],"generation":42,"complete":true}`)
		})

		mux.HandleFunc("GET /v1/watch", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)

			fmt.Fprint(w, "event: \ndata: {\"type\":\"StreamConnected\"}\n\n")
			flusher.Flush()

			fmt.Fprint(w, ": heartbeat\n\n")
			fmt.Fprint(
				w,
				"data: {\"type\":\"WatchProjectAdded\",\"timestamp\":\"2026-11-17T20:00:00Z\",\"path\":\"/p/new-repo\",\"project\":{\"dirName\":\"new-repo\",\"path\":\"/p/new-repo\"}}\n\n",
			)
			flusher.Flush()
		})

		srv = newUnixTestServer(mux)
		DeferCleanup(srv.close)
	})

	It("discovers projects over the unix socket", func() {
		client := discovery.New(srv.socket)

		projects, err := client.Discover(context.Background(), []string{"/home/lars/projects"})
		Expect(err).NotTo(HaveOccurred())
		Expect(projects).To(HaveLen(2))
		Expect(projects[0].DirName).To(Equal("go-finding"))
		Expect(projects[1].RepoName().Get()).To(Equal("go-compliance"))
	})

	It("resolves the socket path from the daemon's env", func() {
		Expect(discovery.ResolveSocketPath()).To(Equal("/run/project-discovery/daemon.sock"))

		Expect(os.Setenv("PROJECT_DISCOVERY_DAEMON_ADDR", "unix:///custom/sock")).To(Succeed())
		DeferCleanup(func() { _ = os.Unsetenv("PROJECT_DISCOVERY_DAEMON_ADDR") })

		Expect(discovery.ResolveSocketPath()).To(Equal("/custom/sock"))
	})

	It("wraps connection failures with guidance", func() {
		_, err := discovery.New("/nonexistent/daemon.sock").Discover(context.Background(), nil)
		Expect(err).To(MatchError(compliance.ErrIO))
		Expect(err.Error()).To(ContainSubstring("is the daemon running"))
	})

	It("streams watch events, skipping heartbeats and handshakes", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		events, err := discovery.New(srv.socket).Watch(ctx)
		Expect(err).NotTo(HaveOccurred())

		var received []discovery.Event
		for event := range events {
			received = append(received, event)
			if len(received) == 2 {
				cancel()
			}
		}

		Expect(received).To(HaveLen(2))
		Expect(received[0].Type).To(Equal("StreamConnected"))
		Expect(received[1].Type).To(Equal("WatchProjectAdded"))
		Expect(received[1].Project.DirName).To(Equal("new-repo"))
	})

	It("names discovered repos for never-covered completion", func() {
		names := discovery.RepoNames([]discovery.Project{
			{DirName: "a"}, {DirName: "b"},
		})
		Expect(names).To(HaveLen(2))

		completed := compliance.CompleteWithDiscovery([]compliance.RepoEvidence{{
			Repo: compliance.NewRepoName("a"), Runs: 1,
		}}, names)
		Expect(completed).To(HaveLen(2))
		Expect(completed[0].Repo.Get()).To(Equal("a"))
		Expect(completed[1].Repo.Get()).To(Equal("b"))
		Expect(compliance.DeriveState(completed[1], time.Now(), compliance.DefaultStaleAfter)).
			To(Equal(compliance.StateNeverCovered))
	})

	It("keeps only projects the fleet producer can cover", func() {
		root := GinkgoTB().TempDir()
		goRepo := filepath.Join(root, "go-repo")
		siteRepo := filepath.Join(root, "site")

		Expect(os.MkdirAll(goRepo, 0o750)).To(Succeed())
		Expect(os.MkdirAll(siteRepo, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(goRepo, "go.mod"), []byte("module x\n"), 0o600)).To(Succeed())

		eligible := discovery.FleetEligible([]discovery.Project{
			{DirName: "go-repo", Path: goRepo},
			{DirName: "site", Path: siteRepo},
			{DirName: "ghost", Path: filepath.Join(root, "missing")},
		})

		Expect(discovery.RepoNames(eligible)).To(Equal([]compliance.RepoName{compliance.NewRepoName("go-repo")}))
	})
})

// syncBuffer serializes writes from the watch goroutine against reads from
// the test goroutine (slog handlers may be called concurrently).
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

var _ = Describe("discovery client SSE framing", func() {
	startWatchServer := func(handler http.HandlerFunc) *unixTestServer {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /v1/watch", handler)

		return newUnixTestServer(mux)
	}

	It("reassembles multi-line data frames into one event", func() {
		srv := startWatchServer(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)

			// Pretty-printed JSON split across data: lines — the spec joins
			// them with "\n" before payload decoding.
			fmt.Fprint(w, "data: {\n")
			fmt.Fprint(w, "data:   \"type\": \"WatchProjectAdded\",\n")
			fmt.Fprint(w, "data:   \"timestamp\": \"2026-11-17T20:00:00Z\",\n")
			fmt.Fprint(w, "data:   \"path\": \"/p/pretty-repo\"\n")
			fmt.Fprint(w, "data: }\n\n")
			flusher.Flush()

			<-r.Context().Done()
		})
		DeferCleanup(srv.close)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventCh, err := discovery.New(srv.socket).Watch(ctx)
		Expect(err).NotTo(HaveOccurred())

		var event discovery.Event
		Eventually(eventCh, "2s", "10ms").Should(Receive(&event))
		Expect(event.Type).To(Equal("WatchProjectAdded"))
		Expect(event.Path).To(Equal("/p/pretty-repo"))

		cancel()
	})

	It("surfaces stream failures instead of faking a clean end", func() {
		srv := startWatchServer(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)

			// One data line over the 1 MiB cap: the parse must fail loudly,
			// not truncate the frame and carry on.
			fmt.Fprint(w, "data: ")
			_, _ = w.Write(bytes.Repeat([]byte("x"), 1536*1024))
			fmt.Fprint(w, "\n\n")
			flusher.Flush()

			<-r.Context().Done()
		})
		DeferCleanup(srv.close)

		logs := &syncBuffer{}
		previous := slog.Default()

		slog.SetDefault(slog.New(slog.NewTextHandler(logs, nil)))
		DeferCleanup(func() { slog.SetDefault(previous) })

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eventCh, err := discovery.New(srv.socket).Watch(ctx)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			select {
			case event, ok := <-eventCh:
				if ok {
					Fail(fmt.Sprintf("no event may be delivered from an over-cap frame, got %#v", event))
				}
			default:
			}

			return strings.Contains(logs.String(), "discovery watch stream failed")
		}, "2s", "10ms").Should(BeTrue())
	})
})
