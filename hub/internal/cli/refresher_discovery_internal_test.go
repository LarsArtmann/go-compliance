package cli

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/discovery"
	"github.com/larsartmann/go-compliance/hub/internal/events"
	"github.com/larsartmann/go-compliance/hub/internal/testsupport"
)

func userHome() string {
	home, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())

	return home
}

var _ = Describe("parseDiscoverPaths", func() {
	DescribeTable("expands, trims, and dedupes",
		func(raw string, want []string) {
			got, err := parseDiscoverPaths(raw)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("default", "~/projects", []string{filepath.Join(userHome(), "projects")}),
		Entry("bare tilde", "~", []string{userHome()}),
		Entry("empty disables", "  ", []string{}),
		Entry("comma list with blanks and a duplicate", "/a, /b ,,/a", []string{"/a", "/b"}),
		Entry("tilde inside a list", "~/x,/y", []string{filepath.Join(userHome(), "x"), "/y"}),
	)
})

var _ = Describe("ServeFlags.DiscoverPaths", func() {
	It("is the --discover-paths flag with the fleet root as its default", func() {
		field, ok := reflect.TypeFor[ServeFlags]().FieldByName("DiscoverPaths")
		Expect(ok).To(BeTrue())
		Expect(field.Tag.Get("flag")).To(Equal("discover-paths"))
		Expect(field.Tag.Get("default")).To(Equal("~/projects"))
	})
})

// stubDaemon mirrors the discovery daemon's contract as far as the hub
// uses it: /v1/discover rejects a request without search paths (400), and
// /v1/watch holds an SSE stream open.
type stubDaemon struct {
	*testsupport.UnixServer

	mu       sync.Mutex
	requests [][]string
}

func (s *stubDaemon) searchPaths() [][]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([][]string(nil), s.requests...)
}

// newStubDaemon starts the stub on a unix socket and answers
// /v1/discover with projects.
func newStubDaemon(projects []discovery.Project) *stubDaemon {
	stub := &stubDaemon{}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/discover", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SearchPaths []string `json:"searchPaths"`
		}

		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil || len(body.SearchPaths) == 0 {
			http.Error(w, `{"error":"search_paths: required"}`, http.StatusBadRequest)

			return
		}

		stub.mu.Lock()
		stub.requests = append(stub.requests, body.SearchPaths)
		stub.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		encoded, _ := json.Marshal(map[string]any{"projects": projects})
		_, _ = w.Write(encoded)
	})

	mux.HandleFunc("GET /v1/watch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(": connected\n\n"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})

	stub.UnixServer = testsupport.NewUnixServer(mux)

	return stub
}

func discoveredNames(refresher *refresher) []compliance.RepoName {
	names := refresher.discovered.Load()
	if names == nil {
		return nil
	}

	return *names
}

var _ = Describe("discovery watch", func() {
	It("sends the configured roots and completes only the fleet-eligible projects", func() {
		root := GinkgoTB().TempDir()
		goRepo := filepath.Join(root, "go-repo")
		siteRepo := filepath.Join(root, "site")

		Expect(os.MkdirAll(goRepo, 0o750)).To(Succeed())
		Expect(os.MkdirAll(siteRepo, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(goRepo, "go.mod"), []byte("module x\n"), 0o600)).To(Succeed())

		stub := newStubDaemon([]discovery.Project{
			{DirName: "go-repo", Path: goRepo},
			{DirName: "site", Path: siteRepo},
		})
		defer stub.Close()

		refresher := newRefresher(events.NewHub(), nil, time.Hour, 24*time.Hour, 3, "")
		refresher.discoverPaths = []string{root}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go runDiscoveryWatch(ctx, discovery.New(stub.SocketPath()), refresher)

		Eventually(refresher.DiscoveryLive, "5s", "20ms").Should(BeTrue())
		Eventually(discoveredNames, "5s", "20ms").WithArguments(refresher).Should(
			Equal([]compliance.RepoName{compliance.NewRepoName("go-repo")}))
		Expect(stub.searchPaths()).To(Equal([][]string{{root}}))
	})

	It("marks the stream idle when the daemon goes away, so the age check takes over", func() {
		stub := newStubDaemon(nil)

		refresher := newRefresher(events.NewHub(), nil, time.Hour, 24*time.Hour, 3, "")
		refresher.discoverPaths = []string{"/fleet"}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go runDiscoveryWatch(ctx, discovery.New(stub.SocketPath()), refresher)

		Eventually(refresher.DiscoveryLive, "5s", "20ms").Should(BeTrue())
		stub.Close()
		Eventually(refresher.DiscoveryLive, "5s", "20ms").Should(BeFalse())
	})

	It("is rejected by a daemon when no roots are configured", func() {
		stub := newStubDaemon(nil)
		defer stub.Close()

		refresher := newRefresher(events.NewHub(), nil, time.Hour, 24*time.Hour, 3, "")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Expect(refresher.refreshDiscovered(ctx, discovery.New(stub.SocketPath()), nil)).NotTo(Succeed())
		Expect(discoveredNames(refresher)).To(BeNil())
	})
})
