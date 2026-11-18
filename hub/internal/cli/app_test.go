package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cli"
)

var _ = Describe("compliance-hub CLI", func() {
	var out, errW bytes.Buffer
	var dir string
	var fixturePath string

	BeforeEach(func() {
		out.Reset()
		errW.Reset()

		var err error
		dir, err = os.MkdirTemp("", "compliance-hub-cli-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })

		report := compliance.BuildV1Report([]compliance.RepoEvidence{{
			Repo: compliance.NewRepoName("hub-test-repo"), Runs: 4, Failures: 1,
			LastRunAt: time.Now().Add(-time.Hour), LastRunVerdict: compliance.VerdictFailed,
			GateFindings: map[string]int{"go-test": 2},
			WorstSteps:   []compliance.StepEvidence{{Step: "go-test", Failures: 1, TotalRuns: 4, FailRate: 0.25}},
		}}, time.Now().Add(-7*24*time.Hour), compliance.DefaultStaleAfter, "test", time.Now())

		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())

		fixturePath = filepath.Join(dir, "compliance.json")
		Expect(os.WriteFile(fixturePath, encoded, 0o644)).To(Succeed())
	})

	runCLI := func(args ...string) error {
		return cli.Run(context.Background(), args, &out, &errW)
	}

	It("report emits a decodable v1 contract", func() {
		Expect(runCLI("report", "--file", fixturePath)).To(Succeed())

		report, err := compliance.DecodeV1Report(out.Bytes())
		Expect(err).NotTo(HaveOccurred())
		Expect(report.SchemaVersion).To(Equal(1))
		Expect(report.Repos).To(HaveLen(1))
		Expect(report.Repos[0].Repo).To(Equal("hub-test-repo"))
		Expect(report.Repos[0].State).To(Equal("failing"))
	})

	It("table renders the human fleet view", func() {
		Expect(runCLI("table", "--file", fixturePath)).To(Succeed())

		Expect(out.String()).To(ContainSubstring("Fleet compliance: 1 repo"))
		Expect(out.String()).To(ContainSubstring("REPO"))
		Expect(out.String()).To(ContainSubstring("hub-test-repo"))
		Expect(out.String()).To(ContainSubstring("failing"))
	})

	It("rejects an unknown source at flag parse with every accepted value named", func() {
		err := runCLI("table", "--source", "psychic", "--file", fixturePath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("psychic"))
		Expect(err.Error()).To(ContainSubstring("file buildflow"))
	})

	It("rejects invalid --since and --stale-after values", func() {
		Expect(runCLI("report", "--since", "soon", "--file", fixturePath)).
			To(MatchError(compliance.ErrValidation))
		Expect(runCLI("report", "--stale-after", "0", "--file", fixturePath)).
			To(MatchError(compliance.ErrValidation))
	})

	It("errors with guidance when the evidence file is missing", func() {
		err := runCLI("table", "--file", filepath.Join(dir, "absent.json"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not exist yet"))
	})

	It("serve listens on the configured address and answers health checks", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		port := freePort()
		base := fmt.Sprintf("http://127.0.0.1:%d", port)

		go func() {
			defer GinkgoRecover()
			err := cli.Run(ctx, []string{"serve", "--addr", fmt.Sprintf("127.0.0.1:%d", port), "--file", fixturePath}, &out, &errW)
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() int {
			resp, err := Get(base + "/healthz")
			if err != nil {
				return 0
			}
			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode
		}, "5s", "50ms").Should(Equal(200))

		for _, path := range []string{"/readyz", "/startupz", "/version", "/", "/events", "/datastar.js"} {
			Eventually(func() int {
				resp, err := Get(base + path)
				if err != nil {
					return 0
				}
				defer func() { _ = resp.Body.Close() }()

				return resp.StatusCode
			}, "5s", "50ms").Should(Equal(200), "path %s", path)
		}

		ready, err := Get(base + "/readyz")
		Expect(err).NotTo(HaveOccurred())
		body, err := io.ReadAll(ready.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(ContainSubstring(`"evidence"`))
		Expect(string(body)).To(ContainSubstring(`"status":"pass"`))
		Expect(ready.Body.Close()).NotTo(HaveOccurred())

		cancel()
	})

	It("serve warns when binding a non-localhost address", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		port := freePort()
		base := fmt.Sprintf("http://127.0.0.1:%d", port)

		go func() {
			defer GinkgoRecover()
			err := cli.Run(ctx, []string{"serve", "--addr", fmt.Sprintf("0.0.0.0:%d", port), "--file", fixturePath}, &out, &errW)
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() int {
			resp, err := Get(base + "/healthz")
			if err != nil {
				return 0
			}
			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode
		}, "5s", "50ms").Should(Equal(200))

		Eventually(errW.String, "2s", "20ms").Should(ContainSubstring("warning: binding non-localhost"))
		Eventually(errW.String, "2s", "20ms").Should(ContainSubstring("docs/decisions/bind-posture.md"))

		cancel()
	})

	It("serve stays silent on the loopback default", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer GinkgoRecover()
			err := cli.Run(ctx, []string{"serve", "--addr", "127.0.0.1:8477", "--file", fixturePath}, &out, &errW)
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() int {
			resp, err := Get("http://127.0.0.1:8477/healthz")
			if err != nil {
				return 0
			}
			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode
		}, "5s", "50ms").Should(Equal(200))

		Consistently(errW.String, "300ms", "50ms").ShouldNot(ContainSubstring("warning"))

		cancel()
	})

	It("serve listens on a unix socket and removes it on shutdown", func() {
		socketPath := filepath.Join(dir, "hub.sock")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer GinkgoRecover()
			err := cli.Run(ctx, []string{"serve", "--addr", "unix:" + socketPath, "--file", fixturePath}, &out, &errW)
			Expect(err).NotTo(HaveOccurred())
		}()

		client := unixClient(socketPath)
		Eventually(func() int {
			resp, err := client.Get("http://unix/healthz")
			if err != nil {
				return 0
			}
			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode
		}, "5s", "50ms").Should(Equal(200))

		Consistently(errW.String, "300ms", "50ms").ShouldNot(ContainSubstring("warning"))

		cancel()

		Eventually(func() bool {
			_, err := os.Stat(socketPath)
			return os.IsNotExist(err)
		}, "5s", "50ms").Should(BeTrue())
	})
})

// freePort asks the kernel for a currently-unused localhost TCP port.
func freePort() int {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	port := listener.Addr().(*net.TCPAddr).Port
	ExpectWithOffset(1, listener.Close()).To(Succeed())
	return port
}

// unixClient dials HTTP through a unix-domain socket.
func unixClient(socketPath string) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}}
}

// Get is a tiny stdlib HTTP GET helper for smoke assertions.
func Get(url string) (*http.Response, error) {
	return http.Get(url) //nolint:noctx // smoke test against a local server
}
