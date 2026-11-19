# go-filewatcher

- Reviewed: 2026-11-19 (group D library-utilization audit)
- Consumer modules + pins: **absent** from both modules (`rg fsnotify|filewatcher` in consumer code → only `hub/go.mod:50` fsnotify v1.10.1 as an *indirect tool-graph* dep of the pinned `templ` generator; nothing in the hub binary watches anything).
- Library reviewed at: `git -C /home/lars/projects/go-filewatcher describe --tags --abbrev=0` → `v2.5.0`; `git log -1` → `26459ba 2026-11-17 chore: bump tooling pins and website dependencies (fleet sweep)`.
- Newest tags: `v2.5.0`, `v2.4.3`, `v2.4.2`, `v2.4.1`, `v2.4.0`. **Tag is current** (fetched fresh). Post-tag `git log v2.5.0..HEAD --oneline` = 6 commits, all docs/chore (release proof, backlog ranking, flake re-lock) — no library `.go` changes.
- Build note: the module requires Go ≥ 1.26.7 with `encoding/json/v2` (README "Install"). The hub module is `go 1.27.2` and already imports stdlib `encoding/json/v2` (hub/internal/cache/cache.go:11) — the requirement is satisfied with no `GOEXPERIMENT` needed on 1.27.

## Verdict: NOT-ADOPTED-but-should

The hub's headline promise is a *realtime* dashboard (AGENTS.md: SSE + DataStar patches, live rows), and its liveness model has a documented hole exactly where this library fits: `refresher` refreshes on (a) this process's cache writes (`store.OnWrite(... refresher.Kick())`, serve.go:59 — but the OnWrite hook "only sees THIS process's writes", AGENTS.md:103), (b) discovery fleet changes, and (c) a **5-minute safety-net ticker** that exists precisely to catch "external processes appending to the same SQLite file" (refresher.go:23-26, 179). Worse, the **default source is `file`** (flags.go:22), for which *no* write trigger exists at all — `OnWrite` is wired only for `*cache.Store` (serve.go:58-60), so a stock `compliance-hub serve` serves evidence that is, by construction, up to five minutes stale no matter how eagerly the producer rewrites the document. go-filewatcher closes both: watch the SQLite DB's parent directory and the -wal/-shm sidecars fire on every cross-process commit (go-sqlitestore runs `PRAGMA journal_mode=WAL`, engine.go:216 of sqlitestore v0.1.1 — in WAL mode commits are `pwrite(2)`s to `-wal`, which inotify reports; the main DB file alone would be nearly silent between checkpoints, so directory-level watching of the sidecars is *the* correct shape); watch the evidence file's directory and the reporter's atomic temp-rename (reporter.go:385) arrives as a Create/Write event on the new name — catching rename-based writes is precisely where watching the *file itself* breaks (inode swap) and watching the *directory* does not. The lib's API matches the need point-for-point: directory roots are the only thing it watches (`validateDirExists`, watcher.go:184-196 — files are rejected with `ErrPathNotDir`, which forces the correct dir-watch pattern), `WithRecursive(false)` + a three-name `Filter` selects `<db>`, `<db>-wal`, `<db>-shm`, `WithSelfHeal` re-arms dropped watches, `Healthy()` is signature-compatible with the hub's go-health `CheckFunc` (lib health.go:6-32, doc comment shows the `health.NewChecks` wiring verbatim), and the retained 5-minute safety net remains the backstop so every failure mode degrades to *today's* behavior, never worse. The cost that keeps this honest: a new direct dep set in the hub (fsnotify — already in go.sum — plus first-party gogenfilter, go-gitignore, x/text, x/time and indirects), a STACK.md foundation row + decision-log entry (stack-drift-check enforces it), and the nix vendorHash bootstrap ritual (AGENTS.md). That is the price of ~150× fresher data (5 min → ~1–2 s) on the product's core feature.

**Flip trigger (adopt when):** the realtime promise is taken seriously for cross-process reporters or the default file source — concretely, when (a) any producer other than the hub process writes the cache DB while `serve` runs (the M27 architecture says reporters do exactly that: the cache is "the reporter SDK's durable sink", cache.go:1-4), or (b) `--source file` (the default!) with a live producer is a real deployment rather than a fallback. **Defer only if** the hub is exclusively used one-shot (`report`/`table`) or the evidence producer legitimately ticks slower than 5 minutes.

## Derived score

- Capability inventory rows: **N = 12**
- N/A rows: 6 → **applicable A = 6**
- Fully = 0, Partial = 0, Missed = 6 → **Σ = 0**
- **Score = 0 of 6 — low band. Publish "0 of 6".** (Utilization is zero because nothing is imported; the verdict carries the should-adopt. Rows 1–3 are the gap itself, not hypothetical capabilities.)

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| Directory watch with child-event delivery (inotify/kqueue; non-recursive mode) | Missed | `watcher.go:369` `Watch(ctx)`; `options.go:81` `WithRecursive` (default true — set false) | `refresher.go:179` — the only cross-process signal is `time.NewTicker(refreshSafetyNet)` (5 min, const at refresher.go:25); AGENTS.md:102-104 documents the limitation |
| SQLite WAL sidecar awareness (watch `-wal`/`-shm`, not just the main DB) | Missed | any-event filter + dir watch delivers child writes; `event.go:99-125` Event carries Path/Op | WAL confirmed at go-sqlitestore v0.1.1 `engine.go:216` (`PRAGMA journal_mode=WAL`); consumer writes cross-process via `cache.Store.WriteEvidence` (cache.go:99-135) from reporter processes |
| Evidence-file live updates for `--source file` (atomic-rename-safe) | Missed | dir-watch catches Create/Write on the renamed-in file (watching the file itself would miss swaps) | flags.go:22-23 default `source=file` + evidence path; serve.go:58-60 wires `OnWrite` **only** for `*cache.Store` → file source has zero write triggers; producer rewrite via rename at reporter.go:385 |
| Event filtering (select DB + sidecars among directory siblings) | Missed | `filter.go:23-30` (`Filter` = `func(Event) bool`); `FilterGlob`, `FilterRegex`, `FilterOperations` | no filtering need today because nothing watches — adoption requires exactly one ~5-line filter |
| Self-healing watch re-registration | Missed | `watcher_selfheal.go:41-100`; `options.go:448` `WithSelfHeal(interval)` | nothing to compare — the 5-min ticker is the consumer's only resilience here |
| Health surface (`Healthy()` as go-health `CheckFunc`) | Missed | `health.go:32` (`Healthy(ctx) error`); doc comment at health.go:6-30 shows `health.NewChecks({"filewatcher": watcher.Healthy})` | hub probe is go-health (`cli/health.go`, STACK.md row); a non-critical `source-watch` check slots into the existing warn-not-fail verdict pattern (cf. `discovery` check) |
| Debounce/coalescing | N/A — the refresher already owns coalescing: cap-1 trigger channel + 500 ms debounce (refreshDebounce const, refresher.go:17-20; loop at refresher.go:192-205); lib debounce (`options.go:24,36`) is optional tuning, not a gap | `options.go:24-45` | refresher.go:192-205 |
| Recursive directory walking | N/A — exactly one directory is watched | `watcher_walk.go`; `options.go:81` | — |
| Polling fallback (NFS/FUSE) | N/A — local ext4/tmpfs deployment; inotify is available (the hub binds `/run` sockets and reads `/var/lib`) | `watcher_poll.go:20-45`; `options.go:232` | — |
| `.gitignore`-aware walking | N/A — no repo tree walked | `watcher_gitignore.go`; `options.go:349` | — |
| Content hashing (`WithContentHashing`) | N/A — the refresher re-derives unconditionally; content identity adds nothing | `options.go:418` | — |
| Middleware / metrics / OTel | N/A — single-operator service, no scraper, no tracer | `middleware.go`, `metrics.go`, `otel.go` | — |

## Gap analysis

### 1. Cross-process cache writes wait up to 5 minutes (Missed — the documented gap)

What the lib offers: events within milliseconds of another process's commit. In WAL mode a committing process appends frames to `<db>-wal` with `pwrite(2)` — an inotify-visible write; the watching process's directory watch delivers it. (Caveat handled by design: `-shm` is updated via mmap and mmap writes do **not** reliably fire inotify — do not count on `-shm` events; `-wal` and the main DB at checkpoint/truncate time carry the signal. Watching the *directory* rather than any single file also survives SQLite's checkpoint cycles, where `-wal` is truncated/reset or deleted and recreated on last-connection-close.)

What the consumer does: `refreshSafetyNet = 5 * time.Minute` (refresher.go:25) with the comment "external processes appending to the same SQLite file get picked up eventually" — a 150–300× latency penalty on the product's core loop (commit → dashboard patch).

Impact: high for the realtime promise, zero for correctness (the data is eventually right). Realistic post-adoption latency: inotify event (~ms) + optional lib debounce + `Kick()` + 500 ms refresher debounce (refresher.go:192-205) + one `QueryEvidence` (local SQLite, ms) ≈ **1–2 s end-to-end**.

### 2. The default `--source file` has no live trigger at all (Missed — the undocumented aggravation)

`ServeFlags` inherit `Source default:"file"` (flags.go:22) and `File default:"/var/lib/overview/compliance.json"` (flags.go:23). serve.go:58-60 wires the write hook only `if store, ok := source.(*cache.Store)` — a `FileSource` never matches, so its *only* refresh signals are discovery fleet changes and the 5-minute ticker. The producer recipe rewrites the document via the reporter's temp+rename (reporter.go:362-386) — invisible to any file-content mtime heuristic that assumed in-place writes, but a textbook Create event for a directory watcher. Same fix, same watcher, one more filter entry.

### 3. No watch-health visibility (Missed)

If a watch silently dies today's design has no signal — the fresh-data promise erodes to the safety net with nothing in `/readyz` saying so. The lib ships the exact probe: `Healthy(ctx)` (health.go:32) returns a joined, `errors.Is`-able error covering closed/not-watching/circuit-open/failed-paths (health.go:6-60) and its doc comment shows the go-health wiring verbatim. The hub's non-critical warn-not-fail pattern (`discovery` check, AGENTS.md health surface) is the established slot for it.

### Integration sketch (hub/internal/cli/refresher.go + serve.go)

```go
// refresher.go — one new exported-ish helper next to runDiscoveryWatch:

// runSourceWatch watches the source's on-disk location and kicks the
// refresher when ANOTHER process changes it (the OnWrite hook only sees
// this process's writes — AGENTS.md "Live Architecture" limitation). The
// 5-minute safety net stays as the backstop for dropped watches.
func runSourceWatch(ctx context.Context, watchedPath string, refresh *refresher) {
	dir := filepath.Dir(watchedPath)
	base := filepath.Base(watchedPath) // e.g. "compliance.db" or "compliance.json"

	w, err := filewatcher.New([]string{dir},
		filewatcher.WithRecursive(false),           // one directory, no walk
		filewatcher.WithGitignore(false),          // /var/lib is not a repo
		filewatcher.WithFilter(func(e filewatcher.Event) bool {
			b := filepath.Base(e.Path)
			return b == base || b == base+"-wal" || b == base+"-shm" ||
				strings.HasPrefix(b, "."+base+".") // the reporter's ".compliance.json.*" temps (optional)
		}),
		filewatcher.WithSelfHeal(30*time.Second),  // re-arm dropped watches
		filewatcher.WithErrorHandler(func(err error) {
			slog.Warn("compliance-hub: source watch error", "error", err)
		}),
	)
	if err != nil {
		slog.Warn("compliance-hub: source watch unavailable; safety net only", "error", err)
		return // degraded = today's behavior, never fatal (serve boot contract stays on Refresh)
	}
	defer func() { _ = w.Close() }()

	events, err := w.Watch(ctx)
	if err != nil {
		slog.Warn("compliance-hub: source watch start failed; safety net only", "error", err)
		return
	}
	slog.Info("compliance-hub: source watch armed", "dir", dir)

	for range events {
		refresh.Kick() // refresher's cap-1 channel + 500ms debounce does the coalescing
	}
}

// serve.go — after resolveSource, before go refresher.Run(ctx):
if watched := watchedPathFor(flags); watched != "" { // db path for cache, File for file source
	go runSourceWatch(ctx, watched, refresher)
}
```

Plus one non-critical probe check (`"source-watch": watcher.Healthy` per lib health.go:32) in `buildHealthProbe`, mirroring the `discovery` check's warn semantics. Keep `refreshSafetyNet` exactly as is (refresher.go:25) — the watcher adds triggers; the ticker remains the floor.

Testing note: the M27 E2E shape (`refresher_internal_test.go`: store write → OnWrite → Kick → debounce → refresh → PublishRefresh) extends naturally — an external-process write can be simulated by writing the DB from a second `cache.Open` handle in-process... no: use a second *process* (or `os/exec` helper) to honor the limitation's letter; alternatively assert on the watcher layer separately (fake backend exists: `fake_backend_test.go` pattern in the lib) and keep the M27 test as-is.

Failure modes, honestly: (1) **watch drop / inotify overflow** — self-heal re-arms; if it can't, safety net still bounds staleness at 5 min and `Healthy()` surfaces it in `/readyz`; (2) **editor-style atomic renames** — a non-issue by construction (dir-watch sees the rename-in as Create; this is also why file-watching was rejected above); (3) **`-wal` deleted on last-writer close** — Remove event fires, filter passes it, one redundant refresh, harmless; (4) **dir missing at Watch() time** — `validateDirExists` errors (watcher.go:184-196): log-and-degrade (sketch above), because the boot contract already gates on `refresher.Refresh` succeeding, not on the watcher; (5) **dep/policy friction** — fsnotify enters the hub's *build* graph for real (today it is only the templ tool's), gogenfilter/go-gitignore/x/text/x/time land as direct requires → STACK.md row + decision-log entry + `stack-drift-check.sh` re-run + nix vendorHash bootstrap (two flake sites + templ-gen devShell per AGENTS.md).

## Version currency

`v2.5.0` = newest tag; post-tag history is docs/chore only. fsnotify pinned by the lib (v1.10.1) is byte-identical to the version already in `hub/go.sum:54` — no version friction. The module's Go floor (1.26.7) is below the hub's (1.27.2); the `jsonv2` requirement is satisfied by the stdlib GA on 1.27 (and the hub already imports it, cache.go:11).

## Fleet leverage

**Adopt — this is the fleet constraint working as intended.** Watch plumbing (debounce, filters, self-heal, error channels, health, overflow resilience) is *exactly* the class of infrastructure every project re-implements badly once and debugs forever; the lib already carries the -race-tested implementations and a health seam that matches this fleet's go-health standard. The counterweight is real but bounded: five direct deps (one first-party), a STACK row, and a vendorHash ritual — priced against a 150× freshness improvement on the hub's headline feature and the retirement of a documented limitation. The alternative "cheap" fix (hand-rolled fsnotify loop or `PRAGMA data_version` polling) would re-create debounce/self-heal/health piecemeal — the precise anti-pattern the owner's fleet constraint forbids — and `data_version` polling still needs its own goroutine, interval tuning, and health story while remaining 30–60× slower than inotify at equal effort.

## Ranked opportunities

1. **Adopt for `--source cache`: watch the DB dir, filter `db`/`-wal`/`-shm`, Kick on event; keep the safety net; add `source-watch` non-critical check** (impact: high — closes the documented 5-min cross-process hole; ease: moderate — the sketch above is ~50 lines + STACK/venderHash rituals).
2. **Adopt for `--source file`: same watcher on the evidence document's dir** (impact: high — gives the *default* source its first live trigger; ease: trivial once (1) lands — one more filter name, one `watchedPathFor` branch).
3. **Fold `refreshSafetyNet` const comment into the new reality** (impact: small — docs truthfulness; ease: trivial; AGENTS.md "Live Architecture" section must be updated with the adoption).

## How to verify

```bash
# Library state
git -C /home/lars/projects/go-filewatcher fetch --tags && git -C /home/lars/projects/go-filewatcher describe --tags --abbrev=0   # v2.5.0
git -C /home/lars/projects/go-filewatcher log -1 --format='%h %cs %s'      # 26459ba 2026-11-17 chore
git -C /home/lars/projects/go-filewatcher log v2.5.0..HEAD --oneline       # 6 docs/chore commits
rg -n 'WithSelfHeal|WithRecursive|WithPolling|WithGitignore' /home/lars/projects/go-filewatcher/options.go     # :448 :81 :232 :349
rg -n 'func \(w \*Watcher\) Watch|validateDirExists|func \(w \*Watcher\) Healthy' /home/lars/projects/go-filewatcher/watcher.go /home/lars/projects/go-filewatcher/health.go  # watcher.go:369, :184; health.go:32

# The gap in the consumer (all verified 2026-11-19)
cd /home/lars/projects/go-compliance
rg -n 'refreshSafetyNet|safetyNet :=' hub/internal/cli/refresher.go        # :25, :179
rg -n 'OnWrite|cache.Store' hub/internal/cli/serve.go                      # :58-60 (cache.Store only)
rg -n 'source.*default:"file"|File .*default' hub/internal/cli/flags.go    # :22, :23
rg -n 'journal_mode=WAL' /home/lars/.cache/go-mod/github.com/larsartmann/go-sqlitestore@v0.1.1/engine.go   # :216
rg -n 'fsnotify|filewatcher' -g '!vendor/**' .                              # hub/go.mod:50 (tool-graph only) — no runtime usage
rg -n 'json/v2' hub/internal/cache/cache.go                                 # :11 (jsonv2 already in use)

# WAL-sidecar firing sanity (optional, ~5 s, no code changes): watch a scratch DB dir while a second process writes
# d=$(mktemp -d); python3 -c "import sqlite3;c=sqlite3.connect('$d/x.db');c.execute('pragma journal_mode=wal');c.commit()"
# inotifywait -m -e modify,create,move "$d" &  # then: python3 -c "import sqlite3;c=sqlite3.connect('$d/x.db');c.execute('create table if not exists t(v)');c.execute('insert into t values(1)');c.commit()"
# → observe MODIFY on x.db-wal (not x.db) — the sidecar claim above
```

All read-only commands executed 2026-11-19 with outputs as annotated; the optional inotifywait demo was not executed (inotifywait not installed on this machine) — the WAL-write-fires-`-wal`-events claim rests on SQLite's documented WAL commit path (frames appended via `pwrite(2)` to the `-wal` file) plus inotify's `IN_MODIFY` on `write(2)`-family syscalls, and on the lib's directory-watch delivering child events.

---

## Addendum 2026-11-19 (executed): adopted

Post-audit Wave 2 (plan `docs/planning/2026-11-19_07-55_*`, T3) flipped this
report's lone NOT-ADOPTED-but-should verdict: `hub/internal/cli/watcher.go`
now dir-watches the cache DB + `-wal`/`-shm` siblings (and the evidence
file's directory for `--source file`, catching atomic renames) and kicks the
refresher — cross-process SQLite writes go from a 5-minute safety-net wait
to ~1-2s. v2.5.0 (public, `/v2` module path — no FLAKE_PAT impact); STACK.md
foundation row + decision-log entry added; the 5-min safety net stays as the
backstop; `watcher.Healthy` feeds a non-critical `evidence-watch` health
check. Pinned by `watcher_internal_test.go`: a write through a second,
independent cache connection (the cross-process stand-in, invisible to
OnWrite) refreshes the hub within the 2s budget, under `-race`. Used
surface: `New`, `Watch`, `WithRecursive(false)`, `WithDebounce`,
`WithFilter`, `Healthy`. Score moves 0/6 → ADOPTED; the report's flip
trigger is discharged.
