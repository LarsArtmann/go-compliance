// Package discovery talks to the project-discovery-daemon over its unix
// socket with plain net/http: no SDK dependency (the SDK is private; this
// repo is public). The daemon owns fleet repository discovery; the hub
// owns turning "repos that exist" into never-covered evidence rows.
package discovery

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/larsartmann/go-compliance"
)

// DefaultSocketPath is the daemon's filesystem socket; the daemon's own
// env override wins when set.
const (
	defaultSocketPath = "/run/project-discovery/daemon.sock"
	envDaemonAddr     = "PROJECT_DISCOVERY_DAEMON_ADDR"

	discoverTimeout = 60 * time.Second
)

// ResolveSocketPath returns the configured daemon socket path.
func ResolveSocketPath() string {
	if fromEnv := os.Getenv(envDaemonAddr); fromEnv != "" {
		return strings.TrimPrefix(fromEnv, "unix://")
	}

	return defaultSocketPath
}

// Project is the discovery wire shape the hub needs: the directory name
// (the canonical fleet repository identifier, matching BuildFlow's
// repo_name) and the filesystem path.
type Project struct {
	DirName string `json:"dirName"`
	Path    string `json:"path"`
}

// RepoName brands the project's directory name as a compliance repo name.
func (p Project) RepoName() compliance.RepoName { return compliance.NewRepoName(p.DirName) }

// discoverResponse mirrors POST /v1/discover (the fields the hub reads).
type discoverResponse struct {
	Projects []Project `json:"projects"`
}

// Event is one /v1/watch SSE event (the daemon's domain.Event shape).
type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Project   *Project  `json:"project"`
}

// Client is the daemon HTTP client over the unix socket.
type Client struct {
	socketPath string
	http       *http.Client
}

// New returns a Client for the given socket path.
func New(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		http: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer

					return d.DialContext(ctx, "unix", socketPath)
				},
			},
		},
	}
}

// SocketPath returns the configured path.
func (c *Client) SocketPath() string { return c.socketPath }

// Discover returns the daemon's current project snapshot for the search
// paths.
func (c *Client) Discover(ctx context.Context, searchPaths []string) ([]Project, error) {
	ctx, cancel := context.WithTimeout(ctx, discoverTimeout)
	defer cancel()

	body, err := json.Marshal(map[string]any{"searchPaths": searchPaths})
	if err != nil {
		return nil, compliance.NewInternalError(err, "encode discover request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://daemon/v1/discover", bytes.NewReader(body))
	if err != nil {
		return nil, compliance.NewInternalError(err, "build discover request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, compliance.NewIOError(err, "discover via %s (is the daemon running?)", c.socketPath)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, compliance.NewIOError(fmt.Errorf("daemon returned %s", resp.Status),
			"discover via %s", c.socketPath)
	}

	var decoded discoverResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, compliance.NewIOError(err, "read discover response")
	}

	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return nil, compliance.NewParseError(err, "decode discover response")
	}

	return decoded.Projects, nil
}

// RepoNames returns the discovered projects' repository names.
func RepoNames(projects []Project) []compliance.RepoName {
	names := make([]compliance.RepoName, 0, len(projects))
	for _, project := range projects {
		names = append(names, project.RepoName())
	}

	return names
}

// Watch subscribes to the daemon's /v1/watch SSE stream and delivers
// decoded events until ctx is cancelled. The returned channel closes when
// the stream ends; callers treat a closed stream as "re-subscribe with
// backoff" (M27 wires the re-derive trigger).
func (c *Client) Watch(ctx context.Context) (<-chan Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://daemon/v1/watch", nil)
	if err != nil {
		return nil, compliance.NewInternalError(err, "build watch request")
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, compliance.NewIOError(err, "watch via %s (is the daemon running?)", c.socketPath)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()

		return nil, compliance.NewIOError(fmt.Errorf("daemon returned %s", resp.Status),
			"watch via %s", c.socketPath)
	}

	events := make(chan Event)

	go func() {
		defer close(events)
		defer func() { _ = resp.Body.Close() }()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue // event names, comments (heartbeats), blank lines
			}

			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "" {
				continue
			}

			var event Event
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				continue // a malformed frame never kills the stream
			}

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}
