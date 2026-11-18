package cli

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// unixDialClient dials HTTP through a unix-domain socket (internal variant).
func unixDialClient(socketPath string) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer

			return d.DialContext(ctx, "unix", socketPath)
		},
	}}
}

// discardLogger keeps serveListener's sd_notify logging out of test output.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

var _ = Describe("bind address handling", func() {
	DescribeTable("isLoopbackBind",
		func(address string, expected bool) {
			Expect(isLoopbackBind(address)).To(Equal(expected))
		},
		Entry("named localhost", "localhost:8477", true),
		Entry("IPv4 loopback", "127.0.0.1:8477", true),
		Entry("IPv6 loopback", "[::1]:8477", true),
		Entry("unix socket is same-machine by construction", "unix:/run/compliance-hub.sock", true),
		Entry("all-interfaces bind", "0.0.0.0:8477", false),
		Entry("empty host binds every interface", ":8477", false),
		Entry("public hostname", "example.com:8477", false),
		Entry("public IP", "192.168.1.10:8477", false),
		Entry("garbage is not loopback", "not-an-address", false),
	)

	Describe("bindAddress unix mode", func() {
		var socketPath string

		BeforeEach(func() {
			socketPath = filepath.Join(GinkgoTB().TempDir(), "nested", "hub.sock")
		})

		It("creates the directory, serves requests, and restricts permissions", func() {
			listener, err := bindAddress(context.Background(), "unix:"+socketPath)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = listener.Close() })

			info, err := os.Stat(socketPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o660)))

			server := &http.Server{Handler: http.NotFoundHandler()}
			go func() { _ = server.Serve(listener) }()

			DeferCleanup(func() { _ = server.Close() })

			client := unixDialClient(socketPath)

			Eventually(func() int {
				resp, err := client.Get("http://unix/healthz")
				if err != nil {
					return 0
				}
				defer func() { _ = resp.Body.Close() }()

				return resp.StatusCode
			}, "2s", "20ms").Should(Equal(http.StatusNotFound))
		})

		It("replaces a stale socket file from a previous run", func() {
			Expect(os.MkdirAll(filepath.Dir(socketPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(socketPath, []byte("stale"), 0o644)).To(Succeed())

			listener, err := bindAddress(context.Background(), "unix:"+socketPath)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = listener.Close() })
		})

		It("refuses a socket path whose directory cannot be created", func() {
			filePath := filepath.Join(GinkgoTB().TempDir(), "occupied")
			Expect(os.WriteFile(filePath, []byte("file"), 0o644)).To(Succeed())

			_, err := bindAddress(context.Background(), "unix:"+filepath.Join(filePath, "hub.sock"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("create socket directory"))
		})
	})

	Describe("serveListener unix cleanup", func() {
		It("removes the socket file when the context cancels", func() {
			socketPath := filepath.Join(GinkgoTB().TempDir(), "hub.sock")

			server := &http.Server{Handler: http.NotFoundHandler()}

			ctx, cancel := context.WithCancel(context.Background())
			served := make(chan error, 1)

			go func() {
				defer GinkgoRecover()

				served <- serveListener(ctx, server, "unix:"+socketPath, discardLogger())
			}()

			client := &http.Client{Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer

					return d.DialContext(ctx, "unix", socketPath)
				},
			}}

			Eventually(func() int {
				resp, err := client.Get("http://unix/healthz")
				if err != nil {
					return 0
				}
				defer func() { _ = resp.Body.Close() }()

				return resp.StatusCode
			}, "2s", "20ms").Should(Equal(http.StatusNotFound))

			cancel()

			select {
			case err := <-served:
				Expect(err).NotTo(HaveOccurred())
			case <-time.After(5 * time.Second):
				Fail("serveListener did not return after context cancel")
			}

			Eventually(func() bool {
				_, err := os.Stat(socketPath)

				return os.IsNotExist(err)
			}, "2s", "20ms").Should(BeTrue())
		})
	})
})
