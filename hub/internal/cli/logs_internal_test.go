package cli

import (
	"bytes"
	"log/slog"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

// captureLogs redirects the default slog logger into a buffer until the
// spec ends; it returns everything logged so far.
func captureLogs() func() string {
	var logs bytes.Buffer

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	DeferCleanup(func() { slog.SetDefault(previous) })

	return func() string {
		return strings.TrimSpace(logs.String())
	}
}
