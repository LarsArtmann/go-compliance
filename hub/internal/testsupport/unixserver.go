package testsupport

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/gomega"
)

// unixServerHeaderTimeout bounds slow header stalls on the stub, like the
// real hub.
const unixServerHeaderTimeout = 5 * time.Second

// UnixServer is an HTTP server bound to a unix socket in a fresh temp
// directory, the way the project-discovery daemon serves; specs use it to
// drive real unix-socket clients against a stub handler.
type UnixServer struct {
	listener net.Listener
	server   *http.Server
	socket   string
}

// NewUnixServer starts handler on a unix socket and fails the running spec
// when the socket cannot be created.
func NewUnixServer(handler http.Handler) *UnixServer {
	dir, err := os.MkdirTemp("", "compliance-unix-server-*")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	socket := filepath.Join(dir, "daemon.sock")
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socket)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	server := &http.Server{Handler: handler, ReadHeaderTimeout: unixServerHeaderTimeout}

	go func() { _ = server.Serve(listener) }()

	return &UnixServer{listener: listener, server: server, socket: socket}
}

// SocketPath returns the socket the server listens on.
func (s *UnixServer) SocketPath() string { return s.socket }

// Close stops the server and removes the temp directory.
func (s *UnixServer) Close() {
	_ = s.server.Close()
	_ = os.RemoveAll(filepath.Dir(s.socket))
}
