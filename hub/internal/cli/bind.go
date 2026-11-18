package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LarsArtmann/go-daemon"
)

// unixAddrPrefix selects unix-socket listen mode.
const unixAddrPrefix = "unix:"

const (
	socketDirPerm  = 0o755
	socketFilePerm = 0o660
	// shutdownGracePeriod bounds graceful HTTP drain after ctx cancels.
	shutdownGracePeriod = 5 * time.Second
)

// bindAddress resolves the --addr flag into a listener. Unix-socket
// addresses ("unix:/path.sock") create their directory, remove a stale
// socket file before binding, and restrict permissions to owner+group.
func bindAddress(ctx context.Context, address string) (net.Listener, error) {
	if !strings.HasPrefix(address, unixAddrPrefix) {
		listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
		if err != nil {
			return nil, fmt.Errorf("listen on tcp %s: %w", address, err)
		}

		return listener, nil
	}

	socketPath := strings.TrimPrefix(address, unixAddrPrefix)
	if err := os.MkdirAll(filepath.Dir(socketPath), socketDirPerm); err != nil {
		return nil, fmt.Errorf("create socket directory %s: %w", filepath.Dir(socketPath), err)
	}

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket %s: %w", socketPath, err)
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("listen on unix socket %s: %w", socketPath, err)
	}

	if err := os.Chmod(socketPath, socketFilePerm); err != nil {
		_ = listener.Close()

		return nil, fmt.Errorf("restrict socket %s permissions: %w", socketPath, err)
	}

	return listener, nil
}

// isLoopbackBind reports whether the address binds a loopback interface
// only (localhost or 127.x over TCP, or any unix socket — same machine by
// construction). An empty host (":8477") binds every interface and counts
// as non-loopback.
func isLoopbackBind(address string) bool {
	if strings.HasPrefix(address, unixAddrPrefix) {
		return true
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	if host == "" {
		return false // ":8477" binds every interface
	}

	if host == "localhost" {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

// serveListener binds, notifies systemd, and serves until ctx cancels.
// Listener-first ordering means NotifyReady fires only once the socket
// accepts traffic; under a systemd WatchdogSec the watchdog loop keeps the
// unit from being killed mid-serve. Unix sockets are removed on shutdown.
func serveListener(
	ctx context.Context, server *http.Server, address string, logger *slog.Logger,
) error {
	listener, err := bindAddress(ctx, address)
	if err != nil {
		return err
	}

	daemon.NotifyReady(logger)

	if interval, ok := daemon.WatchdogInterval(logger); ok {
		go daemon.Watchdog(ctx, logger, interval)
	}

	serveError := make(chan error, 1)
	go func() { serveError <- server.Serve(listener) }()

	select {
	case <-ctx.Done():
		daemon.NotifyStopping(logger)

		// Background, not ctx: the parent is already cancelled, and the
		// graceful-drain window must not inherit that cancellation.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
		defer cancel()

		_ = server.Shutdown(shutdownCtx) //nolint:contextcheck // drain window must not inherit parent cancellation

		if unixListener, ok := listener.(*net.UnixListener); ok {
			_ = os.Remove(unixListener.Addr().String())
		}

		return nil
	case err := <-serveError:
		return err
	}
}
