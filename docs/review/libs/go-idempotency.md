# go-idempotency

| Field            | Value                                                                                                                                                                                                                                                                                                                                                 |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19                                                                                                                                                                                                                                                                                                                                            |
| Group            | C (library-utilization audit, HTTP/wire-protocol family)                                                                                                                                                                                                                                                                                              |
| Consumer modules | core — **absent**; hub — **absent** (not in either `go.mod`, direct or indirect)                                                                                                                                                                                                                                                                      |
| Library version  | newest tag overall `idemtest/v0.1.1` (sub-module); **main-module tags**: `v0.6.0` (newest), then v0.5.0, v0.4.0, v0.3.2, v0.3.1, v0.3.0, v0.2.0, v0.1.2, v0.1.1, v0.1.0 (complete list, `git tag \| grep -v idemtest`) — so the main module sits at ~v0.6.x. Last commit `80c3129 2026-11-18 "chore: auto-commit 3 changed file(s) (heuristic)"`      |
| Modules          | root `github.com/larsartmann/go-idempotency` (Store interface, MemoryStore — deprecated, dev/test only; SDK philosophy: no production backends shipped); `github.com/larsartmann/go-idempotency/idemtest` @ v0.1.1 (contract test suite). `middleware/` is a package of the root module (HTTP + Command wrappers). Only runtime dep: go-error-family. |

## What write surface does the hub actually expose? (verified, not assumed)

**None over HTTP.** Every registered route is GET (rg over all `mux.Handle*`
calls in the hub: only `hub/internal/dashboard/routes.go:58-60` and
`hub/internal/cli/serve.go:83`; the discovery client's POST at
`hub/internal/discovery/client.go:110` is _outbound_, not served). Writes flow
through two non-HTTP paths:

1. **The EvidenceWriter port** (`reporter.go:48-50` in core): reporters call
   the one-liner `compliance.Record(...)` (reporter.go:63-70) which writes to
   the fleet JSON document via `FileStore.WriteEvidence` — a locked
   read-modify-write with atomic rename (reporter.go:151-182).
2. **The SQLite cache adapter** (`hub/internal/cache/cache.go:101-138`):
   `Store.WriteEvidence` is a plain append-only INSERT into `evidence_runs`.
   This is what reporters hit when configured with `--db` (flags.go:56 opens
   the cache; the OnWrite hook kicks the refresher, serve.go:58-60).

## Is there duplicate-submission risk? (analyzed per path)

- **File path**: `WriteEvidence` replaces the repo's row "when the result is
  at least as new as the stored last run; older results are ignored (parallel
  reporters cannot regress the document)" (reporter.go:59-62). Re-submitting
  the same run is a no-op or an identical overwrite — **last-writer-wins by
  timestamp is already idempotent for equal payloads**. The fleet lock
  (reporter.go:152-156) serializes concurrent reporters.
- **SQLite path**: append-only, and reads derive "each repo's all-history
  newest record in ONE query" (`queryLatestRecords`, queries.go:60-63 comment;
  "out-of-order records cannot corrupt the view", cache.go:98-100). A
  duplicated INSERT yields two identical rows of which exactly one (the
  newest) is ever read. Harmless by design; the only cost is row bloat.
- **HTTP middleware reach**: even if one _wanted_ dedup, `middleware.HTTP`
  keys on a request `Idempotency-Key` header (middleware/http.go:30) — there
  is no HTTP handler through which evidence arrives, so the middleware has
  literally nothing to wrap.
- **Duplicate-submission semantics**: go-idempotency exists to make at-least-
  once _side effects_ (charge, emit event, send email) at-most-once. Evidence
  recording has no side effect beyond a state upsert keyed by repo+time; a
  replay neither double-charges nor diverges the view. The risk class does not
  exist on this surface.
- **Single operator**: even the "client retried because ack was lost" story is
  thin — reporters are local CLI/CI one-shots writing to a file or socket-owned
  DB, and the refresher debounces (500 ms, refresher.go:17) whatever burst
  arrives.

## Verdict — **NOT-ADOPTED-justified**

The consumer has no HTTP write surface, and both non-HTTP write paths are
idempotent (LWW-by-timestamp file merge; append-only newest-wins SQL view).
Adopting would mean implementing a `Store` backend (the library ships none by
philosophy) to guard a risk that cannot materialize. This is the strongest
justified-no of the five libraries audited.

**Flip trigger:** the hub ever grows an HTTP write endpoint with a real side
effect — e.g. `POST /v1/evidence` accepting remote/network reporters (a
plausible v2-era feature once bind-posture g1 resolves), a report-submission
API, or anything that emits downstream events (notifications, webhooks). At
that moment: `middleware.HTTP(store, ttl, handler)` over the write route plus a
SQLite-backed `Store` implementation (one conditional upsert — the README
names the primitive), verified with the `idemtest` contract suite
(`github.com/larsartmann/go-idempotency/idemtest`, contract tests for exactly this).

## Derived score (counterfactual: what adopting would cover vs already solved)

Bands: 85-100 excellent · 70-84 good · 50-69 fair · <50 poor.

| Capability                                      | Status                | Library evidence (path)                  | Consumer evidence (path:line)                                                                                                                                                             |
| ----------------------------------------------- | --------------------- | ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| At-most-once processing of commands             | Covered by design — 1 | README "How it works" (`CheckAndRecord`) | writes are timestamp-LWW state upserts: file merge ignores non-newer results (reporter.go:59-62,168-170); SQL reads newest-only (queries.go:60-63)                                        |
| HTTP Idempotency-Key middleware                 | N/A                   | `middleware/http.go:30`                  | zero HTTP write endpoints exist to wrap (`hub/internal/cli/serve.go:75-83`, all GET)                                                                                                      |
| Command-wrapper dedup (`middleware.NewCommand`) | N/A                   | `middleware/middleware.go:50`            | no command dispatch surface in the consumer (adapters write directly)                                                                                                                     |
| Atomic key store interface                      | N/A                   | `store.go` (`Store`)                     | no keys to store — nothing deduplicates because nothing double-executes                                                                                                                   |
| Contract test suite for custom stores           | N/A                   | `idemtest/` module                       | no Store implementation exists to test                                                                                                                                                    |
| Concurrency correctness (atomic CheckAndRecord) | Covered elsewhere — 1 | README "atomic" contract                 | the equivalent race is already closed differently: fleet file lock (reporter.go:152-156), SQLite single-writer + busy-retry (go-sqlitestore), OnWrite hook serialization (cache.go:91-99) |

Inputs: N=6, A=3 (1+1+1), Σ=3.0 → **score = round(100×3.0/3) = 100
(excellent)** as a counterfactual: every need in the library's domain is
already met by the consumer's design, without the library. Publish: **3 of 3**.
(For a NOT-adopted lib this score reads as "no unmet need", not "library in
use" — the verdict above is the operative statement.)

## Gap analysis

No actionable gaps. The one forward-looking sketch, parked behind the flip
trigger:

```go
// IF the hub ever serves POST /v1/evidence (network reporters, v2 era):
// 1) implement the Store against the existing SQLite cache — one conditional upsert:
func (s *Store) CheckAndRecord(ctx context.Context, key string, ttl time.Duration) error {
	res, err := s.db.ExecWrite(ctx,
		`INSERT INTO idempotency_keys (key, expires_at) VALUES (?, ?)
		 ON CONFLICT(key) DO NOTHING`, key, time.Now().Add(ttl))
	if err != nil { return compliance.NewIOError(err, "idempotency check") }
	n, _ := res.RowsAffected()
	if n == 0 { return idempotency.ErrDuplicate }
	return nil
}
// 2) wrap the route:
mux.Handle("POST /v1/evidence", idemMiddleware.HTTP(store, 10*time.Minute, ingestHandler))
// 3) pin the store's contract:
//   idemtest.RunSuite(t, newStoreUnderTest)   // from github.com/larsartmann/go-idempotency/idemtest
```

(The schema row would join the append-only `migrations` list in
`hub/internal/cache/cache.go:26-48` — never edit an applied migration.)

## Version currency

Not pinned. Library healthy and active: main module v0.6.0 with a steady tag
history back to v0.1.0, last commit 2026-11-18, `MemoryStore` deprecated with
a stated future-major removal — a consumer adopting the SDK shape (Store impl

- middleware) is insulated from that deprecation by construction. No currency
  risk while absent.

## Fleet leverage

This library's fleet value concentrates where CQRS/at-least-once delivery
already lives (go-cqrs-lite ecosystems, per its own README). For the
compliance family specifically, the leverage argument only activates if hubs
start accepting network writes: at that point a shared SQLite `Store` impl
could live next to go-sqlitestore and every fleet service with a write API
would reuse it — that would be genuine plumbing-once. Today, adopting here
would be fleet _anti_-leverage: a dep + a backend implementation guarding a
risk the domain model already structurally excludes, which every future reader
must understand and justify. Honest both directions: the library is good at
its job; this consumer's job is not its job.

## Ranked opportunities (impact × ease)

1. **Nothing now.** (The parked sketch above is a v2-era note, not a backlog
   item; it exists so the flip trigger has a concrete shape.)

## How to verify (all commands executed 2026-11-19)

```bash
# Library identity & health
git -C /home/lars/projects/go-idempotency describe --tags --abbrev=0        # → idemtest/v0.1.1 (newest tag OVERALL = submodule)
git -C /home/lars/projects/go-idempotency tag | grep -v idemtest            # → v0.1.0 … v0.6.0 (main module at v0.6.0)
git -C /home/lars/projects/go-idempotency log -1 --format='%h %cs %s'       # → 80c3129 2026-11-18 chore: auto-commit...
cat /home/lars/projects/go-idempotency/go.mod   # deps: go-error-family + rapid (test)
rg -n 'func HTTP|func NewCommand' /home/lars/projects/go-idempotency/middleware/*.go  # → http.go:30, middleware.go:50

# Consumer absence
rg -n 'go-idempotency' /home/lars/projects/go-compliance/go.mod /home/lars/projects/go-compliance/hub/go.mod   # → no matches
rg -n 'idempotency' /home/lars/projects/go-compliance --type go -g '!vendor/**' | head  # → no code references

# The hub is read-only over HTTP (the core claim)
rg -n 'mux.Handle' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'
#   → dashboard/routes.go:58-60 (GET x3) and cli/serve.go:83 (GET /v1/compliance) only
rg -n 'http.MethodPost|http.MethodPut|http.MethodPatch|http.MethodDelete' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'
#   → ONLY discovery/client.go:110 (outbound client call), no served routes

# Write paths and their idempotency-by-design
rg -n 'at least as new|older results are ignored|cannot regress' /home/lars/projects/go-compliance/reporter.go   # → :59-62 doc contract
rg -n 'newest record|out-of-order' /home/lars/projects/go-compliance/hub/internal/cache/cache.go                  # → :98-100, :216-219
rg -n 'INSERT INTO evidence_runs' /home/lars/projects/go-compliance/hub/internal/cache/cache.go                   # → append-only write at :101ff
```

Spot-checks: reporter.go:59-70 and 151-182 read in full (lock + LWW merge);
cache.go:91-138 and 216-219 read (OnWrite hook, append-only INSERT, newest-
per-repo query); serve.go:74-102 read (route inventory). The "POST /v1/discover
is outbound" distinction verified at discovery/client.go:110-118 (client-side
`http.NewRequestWithContext`).

## Citation drift addendum (2026-11-19 post-audit wave)

The `cache.go` queryLatestRecords comment citation (216-219) moved to
`queries.go:60-63` in the cache split (verified by
`scripts/check-citations.sh`).
