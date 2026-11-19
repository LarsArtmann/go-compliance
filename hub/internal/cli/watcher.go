package cli

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/larsartmann/go-filewatcher/v2"
)

// evidenceWatchCheckName is non-critical (verdict g1 family): a dead
// evidence watch degrades freshness to safety-net cadence (still truthful,
// just slower) — warn, never fail readiness.
const evidenceWatchCheckName = "evidence-watch"

// evidenceWatchDebounce bounds the watcher's own event coalescing before
// the refresher's trigger channel coalesces again; small on purpose: the
// refresh itself is the expensive part and already debounced there.
const evidenceWatchDebounce = 200 * time.Millisecond

// watchEvidence starts a directory watch over the evidence locations the
// configured source reads and kicks the refresher on relevant changes,
// closing the 5-minute safety-net staleness window for CROSS-PROCESS
// writes (the cache's OnWrite hook only sees this process): SQLite
// reporters append to the <db>-wal/-shm sidecars, atomic writers rename
// the evidence file into place — both fire directory events within ~1-2s.
//
// Watching the PARENT DIRECTORY (not the file) is load-bearing: an
// atomically replaced main file or a checkpointed database re-points a
// file-level watch, and WAL appends only ever touch the sidecars — the
// directory is the one stable vantage point. Returns nil (and logs) when
// the source has no filesystem truth (buildflow) or the watch cannot
// start: the safety net remains the backstop either way.
func watchEvidence(ctx context.Context, flags QueryFlags, refresh *refresher) *filewatcher.Watcher {
	dirs, names := evidenceWatchTargets(flags)
	if len(dirs) == 0 {
		return nil
	}

	watcher, err := filewatcher.New(dirs,
		filewatcher.WithRecursive(false),
		filewatcher.WithDebounce(evidenceWatchDebounce),
		filewatcher.WithFilter(func(event filewatcher.Event) bool {
			_, keep := names[filepath.Base(event.Path)]

			return keep
		}),
	)
	if err != nil {
		slog.Warn("compliance-hub: evidence watch not started; falling back to safety net", "error", err)

		return nil
	}

	events, err := watcher.Watch(ctx)
	if err != nil {
		slog.Warn("compliance-hub: evidence watch not started; falling back to safety net", "error", err)

		return nil
	}

	go func() {
		for range events {
			refresh.Kick()
		}
	}()

	slog.Info("compliance-hub: evidence watch started", "dirs", dirs, "files", len(names))

	return watcher
}

// evidenceWatchTargets returns the directories to watch and the exact file
// names inside them that mean "new evidence may exist": the evidence file
// itself (source=file), or the cache database plus its WAL siblings
// (source=cache). Directories are deduplicated; relative paths are made
// absolute so the name filter matches against the watcher's absolute
// event paths.
func evidenceWatchTargets(flags QueryFlags) ([]string, map[string]struct{}) {
	names := map[string]struct{}{}
	seen := map[string]struct{}{}
	dirs := []string(nil)

	add := func(file string) {
		if file == "" {
			return
		}

		abs, err := filepath.Abs(file)
		if err != nil {
			return
		}

		if dir := filepath.Dir(abs); dir != "" {
			if _, dup := seen[dir]; !dup {
				seen[dir] = struct{}{}
				dirs = append(dirs, dir)
			}
		}

		names[filepath.Base(abs)] = struct{}{}
	}

	switch flags.Source {
	case "", "file":
		add(flags.File)
	case "cache":
		add(flags.DB)
		add(flags.DB + "-wal")
		add(flags.DB + "-shm")
	}

	return dirs, names
}
