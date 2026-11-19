# go-atomic-write

- Reviewed: 2026-11-19 (group D library-utilization audit)
- Consumer modules + pins: hub `github.com/larsartmann/go-atomic-write v0.6.2` (**indirect** — puller chain verified: `cd hub && GOWORK=off go mod why` → `hub/cmd/compliance-hub` → `cmdguard/v5` → `samber-do-auditlog v0.12.3` → `go-atomic-write`, used at `plugin.go:246` `atomicwrite.WriteFunc` for audit-report export). Core: **absent** (zero private *and* zero extra public deps — root `go.mod` requires only go-branded-id + go-error-family per STACK.md wiring order).
- Library reviewed at: `git -C /home/lars/projects/go-atomic-write describe --tags --abbrev=0` → `v0.6.2`; `git log -1` → `4f10347 2026-11-15 chore: auto-commit 3 changed file(s) (heuristic)`.
- Newest tags: `v0.6.2`, `v0.6.1`, `v0.6.0`, `v0.5.2`, `v0.5.1`. **Pin is current** (tags fetched fresh). Post-tag history adds a real but small API: `WriteFuncWithPerm` + parent-dir `MkdirAll` in `tempPathFor` + a temp-path-collision test contract (`git diff v0.6.2..HEAD -- atomicwrite.go stage.go` — 2 lib files, additive only). Not needed by any consumer path today.

## Verdict: NOT-ADOPTED-justified (indirect only; the one direct-write site is core's, and core's dep policy excludes the lib)

The consumer's entire file-write surface is a single function: core's `writeFileAtomic` (reporter.go:362–386 — `os.CreateTemp` + `Write` + `Chmod` + `Rename`), used by `FileStore.WriteEvidence` (reporter.go:151) to persist the fleet evidence document under a sidecar `flock` read-modify-write protocol (reporter.go:395–411, `lock_unix.go`). Everything else writes through other engines: the hub's SQLite cache goes through go-sqlitestore (its own WAL/durability), the dashboard serves embedded bytes, and the fixture script writes a git-reviewed golden via shell redirect (`scripts/refresh-fixture.sh:28` — crash-safety is the wrong tool for a human-eyeballed regeneration ritual). The hub module itself has **zero** direct file-write sites (`rg 'os\.WriteFile|os\.Create\(|os\.OpenFile|os\.CreateTemp' -g '!vendor' -g '!*_templ.go'` → only reporter.go:400's lock-file open and reporter.go:365's temp create). So the only question is whether *core* should swap its 25-line hand-roll for the lib — and the answer is no under the standing policy: STACK.md fixes core's deps at "go-branded-id + go-error-family only" (module layout section), a deliberate dep-lightness decision for the publishable SDK, and AGENTS.md requires a STACK decision-log entry for new deps. The hand-roll covers the two load-bearing properties for this contract — unique temp + atomic rename (readers never see a torn document; reporter.go:362–386) and cross-process mutual exclusion via `flock` on a sidecar (reporter.go:395) — and the evidence document is *designed* to be re-derivable: each repo re-records on its next run, so even a lost document self-heals. What the hand-roll genuinely lacks vs the lib is **crash durability** (no fsync of the temp file, no fsync of the directory after rename — on ext4, a power loss in the window can leave the renamed file zero-length) and the **fingerprint TOCTOU check** (moot here: the flock protocol already serializes cooperating writers, and a non-cooperating writer would defeat the fingerprint too since the document is rewritten wholesale). That durability gap is real but priced correctly: the fleet evidence document is ephemeral operational metadata, not a ledger.

**Flip trigger:** any of (a) a crash/power-loss ever observed zero-lengthing or reverting `/var/lib/overview/compliance.json` (then the fix is `atomicwrite.WriteVerified` in core's `FileStore` *plus* the STACK decision-log entry — the swap also *deletes* the hand-rolled `writeFileAtomic` + `acquireFileLock` + `lock_unix/other.go`, net code reduction); (b) core's write surface grows a second atomic-write site (a second hand-roll is where the duplication tax starts compounding); (c) the STACK two-dep limit for core is ever revisited (e.g. for other reasons) — go-atomic-write is public, dep-light itself (xxhash + flock + x/sys only), and would ride along cheaply at that point.

## Derived score

- Capability inventory rows: **N = 7**
- N/A rows: 4 → **applicable A = 3**
- Fully = 0, Partial = 0, Missed (hand-rolled equivalents / absent durability) = 3 → **Σ = 0**
- **Score = 0 of 3 — low band. Publish "0 of 3".** (This scores *library utilization*, not consumer correctness: the hand-rolled writer is correct for its contract except crash durability — see gap 1. The verdict carries the justification; the score honestly records that none of the lib's applicable capability is being consumed by the consumer's own code.)

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| Crash-durable atomic write (unique temp → fsync file → rename → fsync dir) | Missed (hand-rolled without the fsyncs) | `atomicwrite.go:17-40` `Write`; `stage.go:73-105` `writeAndSync` (`file.Sync()`); `rename_unix.go:28` directory `Sync()` | `reporter.go:362-386` `writeFileAtomic` — temp+write+chmod+rename, **no** `Sync()` of temp or dir |
| TOCTOU fingerprint verification (`WriteVerified` → `ErrConcurrentModification`) | Missed (different mechanism achieves the cooperating-writer guarantee) | `fingerprint.go:26,35`; `atomicwrite.go:65-90`; `commit.go:13` | `reporter.go:395-411` + `lock_unix.go` — sidecar `flock` exclusive lock held across read-merge-write (flock is cross-process) |
| Cross-platform file locking | Missed (hand-rolled flock via syscall, build-tagged) | `commit.go:36` gofrs/flock (flock/LockFileEx) | `reporter.go:395`; `lock_unix.go` / `lock_other.go` build tags |
| Streaming write (`WriteFunc*`) for large files | N/A — the v1 document is kilobytes (341 repos fixture ≈ 100 KiB class); `MarshalV1` produces `[]byte` anyway | `atomicwrite.go:125-166` | `reporter.go:181-190` |
| `WriteIfChanged` (skip identical content) | N/A — every reporter write changes `generated_at` (reporter.go:257), so "unchanged" never occurs | `atomicwrite.go:91` | `reporter.go:222-258` `mergeRecord` |
| Permission-correct creation (no chmod window) | N/A — consumer chmods the *temp* file before the atomic rename (target never visible with wrong mode); outcome equivalent to `WriteWithPerm` | `atomicwrite.go:42-63` | `reporter.go:380-383` |
| Parent-dir autocreation | N/A — consumer creates dirs at the lock boundary already (`acquireFileLock` → `MkdirAll`, reporter.go:398) | `stage.go:35-43` (post-v0.6.2) | `reporter.go:398` |

## Gap analysis

### 1. No fsync: the durability delta (Missed → the only substantive gap)

What the lib offers: `writeAndSync` fsyncs the staged temp file (stage.go:88) and `commitRename` fsyncs the parent directory after rename (rename_unix.go:28) — the full crash-durability recipe (the README comparison table's "DIY write + rename: crash-durable *Partial*" row describes exactly `writeFileAtomic`).

What the consumer does: `writeFileAtomic` (reporter.go:362–386) renames without either fsync. On ext4 (`data=ordered`), a crash shortly after rename can leave the target zero-length: the rename is journaled but the temp file's data blocks may not be.

Impact: low, by design — the document is the *union of per-repo last runs*, every repo re-records on its next run (reporter.go Record contract), and `FileStore.read` treats a missing/empty file as "start fresh" (reporter.go:195-198). Worst case is a transient fleet-wide `never-covered` blip on the dashboard until reporters cycle, not data loss. Fleet-honest weighing: closing this via the lib costs core its two-dep policy (+xxhash, +gofrs/flock, +x/sys — the latter two already in the *hub* graph, but core must carry them itself); closing it *without* the lib costs two lines:

```go
// reporter.go, inside writeFileAtomic after temp.Close() and before os.Rename:
if f, err := os.Open(temp.Name()); err == nil { _ = f.Sync(); _ = f.Close() }   // data durability
// and after the successful rename:
if d, err := os.Open(dir); err == nil { _ = d.Sync(); _ = d.Close() }          // rename durability
```

That is the honest answer to "does the lib already solve a durability risk here": yes it solves it, but the *solution's* marginal lines (≈4) are smaller than the policy cost of importing it into a public dep-light SDK. Note the `//nolint:wrapcheck` on `writeFileAtomic` (reporter.go:359-360) already marks it as a leaf IO helper — adding Sync calls changes nothing about the wrap posture.

### 2. Fingerprint vs flock (Missed by mechanism, equivalent for this contract)

The lib's `WriteVerified` detects a concurrent modification between read and write via xxhash fingerprint and returns `ErrConcurrentModification` for caller-side retry. The consumer instead *prevents* the race: the sidecar `.lock` flock (reporter.go:395) is held across the whole read-merge-write, and `flock(2)` is cross-process, so every cooperating reporter serializes. A fingerprint would additionally catch a *non-cooperating* writer — but such a writer clobbers wholesale (no merge), which the LWW/newest-run-wins contract (reporter.go:222-246 `mergeRecord`) already tolerates on the next record. No action.

### 3. The indirect puller covers writes that never happen (observation, not a gap)

`go mod why` proves the hub binary links go-atomic-write via cmdguard → samber-do-auditlog (`plugin.go:246`), but the hub never invokes the audit-report export (app.go wires only `WithCLIVersion` + `WithEnvPrefix`), so in practice the lib executes zero writes in this deployment. The pin is correct hygiene (cmdguard requires it); nothing to verify beyond `go mod tidy` staying quiet.

## Version currency

`v0.6.2` is the newest tag; fetched fresh. Post-tag commits (to 2026-11-15) add `WriteFuncWithPerm` and parent-dir autocreation — additive, no consumer need. If the flip trigger ever fires, adopt at a tagged version ≥ v0.6.2; the `WriteVerified` API used in the sketch below is stable since v0.5.x per CHANGELOG.

## Fleet leverage

**Leave.** The fleet already routes this concern through the lib where it belongs (infra libs like samber-do-auditlog embed it; 135-consumer plumbing projects get crash-safe writes for free). Forcing it into go-compliance *core* would invert the fleet's other consolidation decision — the dep-light publishable SDK — to save ~25 lines that are already correct modulo two fsyncs. The poor-fit adoption cost is real here: every downstream importer of core (the public SDK) would inherit xxhash+flock for a durability property its data model explicitly degrades gracefully. The correct fleet move if durability ever matters is the 4-line fsync patch (gap 1), which keeps core dep-neutral. The flip triggers above are genuine, not ceremonial: a second write site or an observed torn document both change the calculus toward adoption.

## Ranked opportunities

1. **Add the two fsync calls to `writeFileAtomic`** (impact: small — closes the only real durability window, keeps core dep-free; ease: trivial, one test with a fault-injection-free assert is enough since fsync failure paths are best-effort `_ =`).
2. **Nothing else.** No consumer path exercises `WriteIfChanged`/`WriteFunc`/perm semantics; do not adopt for completeness.

## How to verify

```bash
# Library state
git -C /home/lars/projects/go-atomic-write fetch --tags && git -C /home/lars/projects/go-atomic-write describe --tags --abbrev=0   # v0.6.2
git -C /home/lars/projects/go-atomic-write log -1 --format='%h %cs %s'                                # 4f10347 2026-11-15 chore
git -C /home/lars/projects/go-atomic-write log v0.6.2..HEAD --oneline | head                          # chores + WriteFuncWithPerm/MkdirAll additions
git -C /home/lars/projects/go-atomic-write diff v0.6.2..HEAD --stat -- '*.go'                         # atomicwrite.go, stage.go + tests

# Indirect puller chain
cd /home/lars/projects/go-compliance/hub && GOWORK=off go mod why github.com/larsartmann/go-atomic-write
# → cmd/compliance-hub → cmdguard/v5 → samber-do-auditlog → go-atomic-write
rg -n 'atomicwrite' /home/lars/.cache/go-mod/github.com/larsartmann/samber-do-auditlog@v0.12.3/plugin.go   # :246 WriteFunc

# Consumer write-surface completeness (the "every file-write site" sweep)
cd /home/lars/projects/go-compliance && rg -n 'os\.WriteFile|os\.Create\(|os\.OpenFile|os\.CreateTemp' -g '!vendor/**' -g '!*_templ.go' .
# → reporter.go:365 (temp), reporter.go:400 (lock file) — non-test, both inside FileStore; scripts/redactfixture writes stdout only
rg -n 'writeFileAtomic|acquireFileLock' reporter.go                                                    # :362, :395
sed -n '355,411p' reporter.go                                                                          # the full hand-roll + lock

# Core dep policy (the adoption constraint)
rg -n 'go-branded-id \+ go-error-family only|Deps: go-branded-id' STACK.md                             # module layout section
rg -n 'Core keeps zero private deps' AGENTS.md
```

All commands executed 2026-11-19; outputs as annotated. File:line citations read directly from the working trees.
