package discovery_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	listener, err := net.Listen("unix", socket)
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
			Expect(jsonNewDecoder(r.Body).Decode(&request)).To(Succeed())
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
			fmt.Fprint(w, "data: {\"type\":\"WatchProjectAdded\",\"timestamp\":\"2026-11-17T20:00:00Z\",\"path\":\"/p/new-repo\",\"project\":{\"dirName\":\"new-repo\",\"path\":\"/p/new-repo\"}}\n\n")
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
})

func jsonNewDecoder(r io.Reader) *jsonDecoder { return newJSONDecoder(r) }
