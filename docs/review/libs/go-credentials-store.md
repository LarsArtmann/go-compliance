# go-credentials-store

- Reviewed: 2026-11-19 (group D library-utilization audit)
- Consumer modules + pins: **absent** from both modules (`rg 'go-credentials-store' go.mod hub/go.mod go.sum hub/go.sum` → no hits; not even transitively).
- Library reviewed at: `git -C /home/lars/projects/go-credentials-store describe --tags --abbrev=0` → `v0.4.3`; `git log -1` → `52b2363 2026-11-18 chore: auto-commit 1 changed file(s) (heuristic)`.
- Newest tags: `v0.4.3`, `v0.4.2`, `v0.4.1`, `v0.4.0`, `v0.2.1`. **Tag is current** (fetched fresh). Post-tag `git log v0.4.3..HEAD --oneline` = 3 chore commits; `git log v0.4.3..HEAD --oneline -- '*.go'` → empty (no code changes).
- What it is (for context): an event-sourced, encrypted-at-rest credentials store (XChaCha20-Poly1305, key rotation, lifecycle audit, loopback admin UI) exposed as `credentials.Build(ctx, opts)` + `container.Resolver().Resolve(ctx, platform, kind)` — built on go-cqrs-lite `system`/`metaengine` with ~10 cqrs sub-module requires plus templ, cmdguard, go-daemon, otter cache (library `go.mod` direct block).

## Verdict: NOT-ADOPTED-justified

The consumer has **no credential surface**, verified end to end: (1) inbound — the hub serves _unauthenticated by decision_ (docs/decisions/bind-posture.md: "No authentication layer until g1 is answered"; default bind 127.0.0.1:8477, unix-socket mode where "socket permissions replace TCP auth entirely", non-localhost binds merely warn, serve.go:110-118); (2) outbound — the hub's only external contacts are the discovery daemon over its unix socket via plain HTTP, no auth (`hub/internal/discovery/client.go:120,168` — the error strings literally say "is the daemon running?", no token anywhere in the file) and the BuildFlow evidence path, which is either a local file read (`FileSource`, adapter.go:27+) or an exec of `buildflow compliance --json` (`NewBuildFlowCommandSource`, adapter.go:190-192 + exec at adapter.go:244) — no BuildFlow state-DB adapter with a connection secret exists in this repo; the SQLite "cache" is the hub's own local file (`cache.Open`, cache.go:66-77); (3) data — evidence is explicitly non-secret by policy: `evidenceFilePerm 0o644` with the comment "fleet evidence is world-readable data, not secrets" (filestore.go:142-144). The one token-adjacent string in the repo is `FLAKE_PAT`, a CI-only GitHub Actions secret for fetching the private go-daemon dep during nix FOD builds (AGENTS.md, "CI hub/flake jobs need the FLAKE_PAT secret") — build-time, never in Go code, never runtime. Adopting an event-sourced encrypted vault (with its cqrs-lite/templ/cmdguard dependency cloud) into a service that holds zero secrets would be the fleet constraint's failure mode, not its win: pure maintenance cost, a second SQLite-adjacent engine in the process, and a bigger attack surface to keep honest for zero capability consumed.

**Flip trigger:** any of (a) open question **g1 lands on "auth"** (bind-posture.md) _and_ the chosen mechanism is per-operator API keys rather than socket-filesystem ownership — at that point the hub needs key storage + rotation and this library is the fleet home for exactly that loop; (b) the hub itself makes an **outbound authenticated call** — e.g. querying Forgejo/GitHub APIs directly instead of letting the discovery daemon own that concern — then `Resolver().Resolve(ctx, "github", KindAPIKey)` is the fleet-standard source (today that need is structurally displaced onto the daemon, STACK.md fleet-discovery row); (c) a remote BuildFlow state-DB adapter (authenticated DSN) is ever added. Until one of those, silence is correct.

## Derived score

- Capability inventory rows: **N = 6**
- N/A rows: 6 → **applicable A = 0**
- **Score: not derivable (A = 0) — publish "n/a (0 of 0)".** There is no capability row where "the consumer should be using this library" is even hypothetically true today; every row is N/A with a reason. A utilization score of 0/0 would be noise, and scoring hypotheticals would inflate the audit's meaning.

## Capability inventory

| Capability                                                            | Status                                                                                                                                                                                                                                                                        | Library evidence (path)                                                                        | Consumer evidence (path:line)                                                                            |
| --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| Secret resolution (`Resolve(ctx, platform, kind)`) for outbound calls | N/A — no outbound call needs a credential: discovery daemon is unauthenticated plain HTTP over unix socket; BuildFlow is local file/exec                                                                                                                                      | `resolve.go`, README "As a library"                                                            | `discovery/client.go:120,168` (no auth headers anywhere in the file); `adapter.go:190-192` (exec, local) |
| Encrypted-at-rest storage + key rotation (KeyID)                      | N/A — nothing secret to store; evidence file is deliberately 0644                                                                                                                                                                                                             | `keys.go`, README "Encrypted at rest by construction"                                          | `filestore.go:142-144` ("world-readable data, not secrets")                                              |
| Inbound auth material (API keys for the admin/hub surface)            | N/A — auth is an open question (g1) and the decided interim is the bind posture: loopback default + unix-socket permissions, no key check                                                                                                                                     | README "non-loopback binds require an API key" (the _library's own server_, not imported here) | `docs/decisions/bind-posture.md` (§ Decision 1-4); serve.go:114-119 non-localhost warning                |
| Lifecycle audit trail (created/rotated/probed/used/revealed)          | N/A — no secrets ⇒ no secret lifecycle to audit; hub auditing is out of scope                                                                                                                                                                                                 | README bullet list                                                                             | —                                                                                                        |
| Loopback admin UI / SSE ops surface                                   | N/A — the hub already _is_ a loopback operator UI; mounting a second one for a zero-secret store is meaningless                                                                                                                                                               | `adminui/` package, README "Exported shim"                                                     | `hub/internal/dashboard` (the hub's own UI)                                                              |
| Event-sourced store engine (cqrs-lite system/metaengine)              | N/A — the hub's persistence is a deliberately simple append-only SQLite table (cache.go:30-58); an event-sourced engine for it was explicitly excluded (STACK.md "Deliberately excluded": go-cqrs-lite — "evidence ingest is adapter writes with no command-side invariants") | library `go.mod` (cqrs-lite requires)                                                          | `cache.go:30-58` (single-table migration); STACK.md excluded table                                       |

## Gap analysis

No Partial/Missed/Misused rows — see score note above. The audit's real work was the credential-surface hunt, so the negative result is recorded with its evidence: `rg -i 'token|authorization|password|secret|credential' -g '!vendor/**' -g '!*_templ.go' -g '!*_test.go' --type go` over both modules returns exactly one hit, the comment at filestore.go:143 asserting the data is _not_ secret (every other hit in the repo is markdown prose about the CI-only `FLAKE_PAT` secret — TODO_LIST.md:47, AGENTS.md:165-166, FEATURES.md, CHANGELOG.md). The FLAKE_PAT reference lives in `.github/workflows` + markdown only (verified by search, not imported by any Go file).

## Version currency

`v0.4.3` is the newest tag; three post-tag chore commits, no `.go` changes. Nothing pinned on the consumer side, so currency is informational only.

## Fleet leverage

**Leave — and treat this row as the template for "justified no."** The fleet constraint cuts both ways: consolidation means not re-implementing plumbing, and it equally means _not importing plumbing you don't need_ — every unused adopter pays upgrade churn, audit surface, and cognitive load that subtract from the collective-maintenance budget the constraint exists to protect. This library has a real fleet role (it is CV's production credential home per its README), and the go-compliance hub is structurally the wrong shape for it: single operator, loopback/unix bind, zero outbound authenticated calls, explicitly non-secret data. The three flip triggers above are the honest boundary; the highest-probability one is (b) — the day the hub wants anything from the Forgejo/GitHub API directly instead of via the discovery daemon, it acquires its first credential and this verdict flips to "adopt the Resolver, not a hand-rolled token file."

## Ranked opportunities

1. **None for adoption.** One preventive note worth a line in docs if g1 is ever reopened: when auth is discussed, decide _key-source_ first (static config key vs socket-ownership vs credentials-store) — the bind-posture escape hatch ("a middleware in front of the mux") presumes a mechanism this library would then supply.

## How to verify

```bash
# Library state
git -C /home/lars/projects/go-credentials-store fetch --tags && git -C /home/lars/projects/go-credentials-store describe --tags --abbrev=0   # v0.4.3
git -C /home/lars/projects/go-credentials-store log -1 --format='%h %cs %s'    # 52b2363 2026-11-18 chore
git -C /home/lars/projects/go-credentials-store log v0.4.3..HEAD --oneline -- '*.go'   # empty (no code changes)

# Absence from consumer
cd /home/lars/projects/go-compliance && rg 'go-credentials-store' go.mod hub/go.mod go.sum hub/go.sum   # no hits

# Credential-surface hunt (the negative result)
rg -in 'token|authorization|password|secret|credential' -g '!vendor/**' -g '!*_templ.go' -g '!*_test.go' --type go .   # filestore.go:143 comment only (Go code)
rg -n 'FLAKE_PAT' -g '!vendor/**' .                              # markdown + .github/workflows only (CI, not runtime; no Go file)
sed -n '1,60p' docs/decisions/bind-posture.md                    # no-auth decision + g1 open question
rg -n 'is the daemon running|http.NewRequest|Header' hub/internal/discovery/client.go   # :113-137, :156-176 — no auth headers
rg -n 'exec.CommandContext' adapter.go                           # :244 — buildflow exec, local
rg -n 'evidenceFilePerm|world-readable' filestore.go            # :142-144
```

All commands executed 2026-11-19; outputs as annotated (the absence hunts return exactly the hits documented, no others).

## Citation drift addendum (2026-11-19 post-audit wave)

The `reporter.go` evidenceFilePerm block (351-353) moved to
`filestore.go:142-144` in the filestore split (verified by
`scripts/check-citations.sh`).

## Citation drift addendum (2026-11-19 post-audit wave)

The `TODO_LIST.md` FLAKE_PAT owner row (line 55 at review time) shifted
with the 2026-11-19 reconcile — it now lives at `TODO_LIST.md:47` in the
Blocked section near the tail (verified by `scripts/check-citations.sh`).
