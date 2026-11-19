# Status Report — go-compliance (2026-11-18 11:21 CET)

Session arc: resumed 07:31 from the 07-19 handoff (M18 flake bootstrap
interrupted mid-apply). Applied the pending flake edit, bootstrapped the hub
vendorHash once, then pivoted into M19 lint work (`.golangci.yml` + core to 0
issues + hub fixes mid-flight) while the second nix attempt ran. This report
was demanded at 11:21 mid-refactor — hub is honestly mid-flight (compile
error, see §d3) and one nix nondeterminism is unresolved (§d4).

## Scope line — what this report verified vs. did not

**Verified with commands this session:**

- `mkPreparedSource` supports `postPatchExtra` at the locked rev a393e23c
  (read from `~/projects/go-nix-helpers`, working tree matches flake.lock).
- Hub FOD attempt 1 printed real hash `sha256-PuP4eGGM…` — proves
  `postPatchExtra` injected the go-daemon replace into `hub/go.mod` and the
  sandbox fetched the private dep.
- Hub FOD attempt 2 (11:21, job output re-read): FAILED with a DIFFERENT
  `got:` hash (`sha256-/8Epx5+…`) — hash flip-flop, not yet root-caused (§d4).
- Hub module compile state (11:21, `GOWORK=off go vet ./...`): 3 errors —
  `bind_internal_test.go` still calls `bindAddress(address)` without ctx.
- Core module (after final core edit, ~07:45): `GOWORK=off go test ./...` ok
  AND `GOWORK=off golangci-lint run ./...` → "0 issues." Core untouched since.
- devShell tools: golangci-lint 2.15.0, templ v0.3.1100, actionlint 1.7.12.
- GitHub mirror is PUBLIC (`gh repo view`); hub needs private go-daemon in CI.
- Git tree clean at 11:21 (auto-commit daemon absorbed all session edits).

**NOT verified (and why):**

- Hub lint count AFTER this session's fixes — re-run never happened
  (interrupted mid-refactor; compile error blocks it anyway).
- Hub test suite — same compile error.
- `nix flake check` — blocked on the vendorHash flip (§d4).
- Root cause of the hash flip — not investigated yet (needs same-rev rebuild
  - vendor-tree diff).
- CI workflow validity — file not written yet (M19 research only).
- All M21+ items — not started.

## a) FULLY DONE (this session)

1. **Todo list resumed** via the `todos` tool exactly per handoff (M17/M20
   completed, M18 in_progress, 10 pending) — kept current through the session.
2. **M18 pending edit applied**: `postPatchExtra` added inside the
   `mkPreparedSource` call in `flake.nix` (edit tool, as instructed; the
   python-heredoc route was correctly avoided). Option verified against the
   pinned helper source before use.
3. **Hub vendorHash bootstrap attempt 1** succeeded at its actual goal: FOD
   reached the go-modules phase and printed the real hash (see scope line).
4. **`.golangci.yml` written** (new file, fleet-style v2 config, deliberately
   leaner than go-finding's — see §e5). This clears the "golangci config"
   quality debt for the core module.
5. **Core module: 0 lint issues, tests green** (both verified). 51 findings
   resolved:
   - Config policy: `tagliatelle json: camel` for hub-facing JSON + exclusion
     `^v1\.go$` (the snake_case v1 wire contract is golden-pinned, not linter
     business); `G302` excluded (evidence files are world-readable fleet
     data, not secrets).
   - Real fixes: `WithMaxOutput(max)` → `limit` (predeclared shadow); consts
     `evidenceFilePerm`/`lockDirPerm`/`maxQueryFlagArgs`; `errors.AsType` in
     errors_test tables; prealloc in reporter_test; `wg.Go` modernize;
     test-fixture hoisting to package scope (fold/state/v1) with the
     eager-Entry-evaluation rationale documented in-file.
   - Documented-contract nolints (each with reason): `read` (nil,nil);
     `resolvedCounters` named 3-tuple; `Rank` rank ladder; wrapcheck at the
     three semantic boundaries (RecordTo delegation, writeFileAtomic/
     acquireFileLock/release syscall helpers, lock_unix wrappers).
6. **M19 research complete**: repo PUBLIC; CI needs the ci-auth pattern
   (PAT + `git config url insteadOf` rewrite — skill reference names
   Code-Quality-Agent/.github/workflows/ci.yml as the fleet implementation);
   drift-check = `stack-selection/scripts/stack-drift-check.sh`; docs-gates =
   `nix run github:Artmann-Technologies/docs-gates -- run .`; lint pin must
   be golangci 2.15.0; actionlint available in devShell.
7. **Hub lint fixes applied (not yet re-verified)**: cache.go (migration wrap,
   `wrapIterateErr` for both `rows.Err()` sites, nolint contracts on count
   marshal/unmarshal); events.go (`NewInternalError` on publish marshal,
   `maps.Copy`); rows.go (`DiffRows` nolint); bind.go (ctx threading through
   `bindAddress`, `net.ListenConfig` for noctx, socket perms + grace period
   consts); serve.go (`buildHealthProbe(ctx,…)`, `readHeaderTimeout` const,
   documented Background-context nolints); health.go (ctx into
   `probe.Start`); colors.go (exhaustive ColorModeAuto case);
   severities.go (rank-map nolint).

## b) PARTIALLY DONE

~~1. **M18 flake bootstrap** — postPatchExtra ✓, attempt-1 hash ✓, pasted ✓;~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
   attempt 2 flip-flopped (§d4). Core-module vendorHash (`sha256-BBBB…`
   placeholder) not started. `nix flake check` not run. M18 closeout
   (TODO_LIST/FEATURES/CHANGELOG updates) not done.
~~2. **M19 CI** — research + config + core-to-zero done; hub fixes applied but~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
   UNVERIFIED; `.github/workflows/ci.yml` not written; `bind_internal_test.go`
   call sites stale (3 compile errors); remaining hub findings untouched
   (discovery client: bodyclose, err113 ×2, mnd 1024; routes.go SplitSeq ×2;
   table_test offBy1; table_internal_test ginkgolinter + sloglint
   DiscardHandler; routes_test stale nolint; app_test noctx; flags.go tagalign
   ×8 + embedded-field order; events.go:228 contextcheck).
~~3. **Quality debt "golangci config"** — cleared for core; hub verdict pending~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
   the re-run.

## c) NOT STARTED (this session)

- M21 examples/reporter + integration guide + README screenshots
- M22 benchmark fold+derive @300 repos + baseline
- M23 BuildFlow TODO_LIST pointer + STACK.md errata
- M24 v2 wire design doc + compat policy
- M25 federation research note
- M27 N+1 fix + store watch → PublishRefresh + Watch client + E2E
- Remaining quality debts: `-race` pass (hub), shared `waitForSubscriber`
  helper, RelativeTime patch purity, AGENTS.md, .gitignore templ note,
  fixture-refresh ritual doc
- M26 (BLOCKED on owner g2 — unchanged)
- Final pass (per-module build+test, plan-vs-tree audit, HARVEST of the 07-19
  report §f — which TODO_LIST.md already absorbed at M17 time)

## d) TOTALLY FUCKED UP

1. **Sloppy regex rename damaged v1_test.go string literals.** A python
   `\bword\b` rename for fixture hoisting rewrote Go STRING CONTENT:
   `"failing-repo"` → `"v1Failing-repo"`, the `"since"` JSON key →
   `"v1Since"`, `By("…failing, stale…")` prose. Caught by post-edit grep,
   fully repaired (round-trip + golden tests then green). Same class as the
   handoff's python-heredoc failure: bulk text tools on Go source keep
   biting; use `edit`/LSP rename.
2. **Hub module currently DOES NOT COMPILE.** The contextcheck fix changed
   `bindAddress` to take `ctx` but the 3 test call sites in
   `bind_internal_test.go` were not updated before the interruption
   (verified: `go vet` errors). Tests were not run after the signature
   change — my own workflow rule broken (test-after-change, immediately).
3. **Hub vendorHash flip-flop (root cause unknown).** Attempt 1 `got:`
   PuP4…, attempt 2 (same go.mod/go.sum) `got:` /8Ep… — the go-modules FOD
   produced two different outputs. Suspects: something version- or
   path-dependent leaking into `vendor/modules.txt` (prepared-source version
   string changed between attempts because the daemon committed mid-boot-
   strap), or nondeterministic vendor output. Needs a controlled experiment:
   build the SAME rev twice, diff the FOD contents (`nix derivation show`,
   compare vendor trees). Do NOT paste-hash-loop past this — the skill's
   gotcha table says flip-flop = investigate.
4. **Background-job discipline.** Nix attempt 2 finished (failed) hours
   before anyone looked — its result surfaced only when this report demanded
   verification. Hanging jobs must be drained at every context switch.

## e) WHAT WE SHOULD IMPROVE

1. Never bulk-regex Go files (again). The two session incidents (heredoc
   syntax error, string-literal damage) are the same failure mode.
2. Run the module suite immediately after ANY signature change — tests are
   call sites; `go vet` is 2 seconds.
3. Drain background jobs (job_output) before starting a new work stream and
   before every report/interruption point.
4. Flip-flopping FOD hashes are a STOP signal, not a loop input — pin the
   rev, rebuild twice, diff.
5. The hand-trimmed `.golangci.yml` (vs go-finding's) dropped: dupl, dupword
   kept… dropped dupl, errname? — the deliberate omissions (dupl, nlreturn,
   wsl_v5, varnamelen, ireturn, goheader, godox, gomoddirectives,
   gochecknoglobals/inits, exhaustruct, spancheck, protogetter, promlinter,
   zerologint, unqueryvet, gosmopolitan, grouper, iface, importas,
   inamedparam, interfacebloat, mirror, sloglint kept…) should be written
   down as policy in AGENTS.md so the trim reads as intent, not as an
   unfinished copy.
6. The ginkgo package-scope-fixture pattern (with the eager-Entry comment)
   is now established in core — apply it in hub test files on sight instead
   of re-deriving per finding.
7. nolint-with-reason is the house pattern for documented-contract findings;
   keep every one carrying its why (all this session's do).

## f) Next — up to 50, impact-ordered

**Unbreak + finish M19 (do these first):**

1. ~~Update the 3 `bindAddress` call sites in `bind_internal_test.go`~~ done — bind_internal_test.go call sites ctx-threaded
   ~~(ctx-threaded signature).~~
2. ~~Run hub tests; fix any fallout from the session's hub edits.~~ done — hub suites green
~~3. Re-run hub lint; fix remaining findings: discovery client bodyclose~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
   (:155), err113 ×2 (sentinel errors), mnd 1024 (read buffer const).
4. ~~routes.go SplitSeq ×2 (modernize stringsseq).~~ done — routes.go SplitSeq
5. ~~table_test.go:73 gocritic offBy1 (Index()+1).~~ done — table_test.go offBy1 guard
6. ~~table_internal_test.go: colors var → package-scope fixture pattern;~~ done — table_internal_test.go package-scope fixture + slog.DiscardHandler
   ~~sloglint → `slog.DiscardHandler`.~~
~~7. routes_test.go:47 remove stale `//nolint:bodyclose`.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
8. ~~app_test.go:229 noctx → `net.ListenConfig`.~~ done — app_test.go net.ListenConfig
~~9. flags.go: tagalign ×8 + embedded-field-before-regular order~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
   (`golangci-lint fmt` likely fixes alignment).
10. ~~events.go:228 contextcheck (r.Context or documented nolint).~~ done — events.go contextcheck nolint with reason
~~11. Hub lint → 0, hub tests green, `-count=1` both modules.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~12. Write `.github/workflows/ci.yml`: per-module build+test (GOWORK=off),~~ done — ci.yml live (per-module build+test, lint, dprint, drift-check, docs-gates, flake-check jobs)
    lint job pinned golangci-lint v2.15.0, dprint job, stack-drift-check job,
    docs-gates job; PAT-auth step guarded on secret presence (ci-auth
    pattern); Forgejo+GitHub-valid constructs only.
~~13. `actionlint` the workflow (devShell has 1.7.12).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~14. Document FLAKE_PAT secret requirement (README CI note + AGENTS.md).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)

**M18 finish:**

~~15. Root-cause the vendorHash flip: build same rev twice, diff FOD vendor~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    trees; likely prepared-source version/path leakage into modules.txt.
~~16. Fix per root cause (possibly go-nix-helpers upstream — ties to §g1).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~17. Complete hub vendorHash; bootstrap core `checks.core-build` hash~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    (`sha256-BBBB…`).
~~18. `nix flake check` green; re-run both module suites.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
19. M18 closeout: TODO_LIST rows removed, FEATURES infra rows flipped,
    CHANGELOG entry; mark M18 completed in todos.

**Then the remaining plan (M21–M27, quality debts, final pass):**

20. ~~M21/F094: `examples/reporter/main.go` using `compliance.Record`.~~ done — examples/reporter/main.go
~~21. M21/F095: `docs/guides/integrate.md` — one line per project.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~22. M21/F096: README screenshots (terminal table + WebUI from fixture).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~23. M22/F097: benchmark fold+derive at 300 repos.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~24. M22/F098: commit baseline under `benchmarks/`.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~25. M23/F099: BuildFlow repo TODO_LIST pointer row + link.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~26. M23/F100: STACK.md errata pass (templ pin note, go-sqlitestore row).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~27. M24/F101: `docs/decisions/` v2 wire design doc (diff vs v1).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~28. M24/F102: compat policy — v1 emittable forever.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~29. M25/F103: federation research (hub ↔ go-health-dashboard).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
30. ~~M25/F104: `docs/notes/federation-interop.md`.~~ done — docs/notes/federation-interop.md
~~31. M27/F107: fix `queryLatestRunCounts` N+1 (cache.go:221, single window~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    query).
~~32. M27/F108: store write-hook → `hub.PublishRefresh`; discovery~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    `/v1/watch` + `Client.Watch` with debounce/coalesce.
~~33. M27/F109: E2E evidence-change → browser-event test (datastartest).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~34. Quality: `go test -race ./internal/...` (hub) green.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~35. Quality: shared `waitForSubscriber` test helper dedupe.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~36. Quality: RelativeTime patch-purity check.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~37. Quality: AGENTS.md for this repo (hub conventions, lint policy incl.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    §e5 trim rationale, verified API facts, ginkgo fixture pattern).
~~38. Quality: .gitignore templ generated-file note.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
39. ~~Quality: fixture-refresh ritual doc (scripts/refresh-fixture.sh usage).~~ done — scripts/refresh-fixture.sh ritual header
~~40. Final pass: per-module `GOWORK=off` build+test both modules.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
41. ~~Final pass: plan-vs-tree audit (every F001–F109 accounted).~~ done — audited 2026-11-18 — 106 verified / 3 owner-blocked (F075 Forgejo visibility, F105/F106 M26) / 0 missing
42. ~~Final pass: docs-health HARVEST of THIS report's §f into TODO_LIST.md.~~ done — this harvest (2026-11-18 final pass)
~~43. Final pass: CHANGELOG Unreleased section complete for the whole train.~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)

**Longer/conditional:**

44. M26 overview endgame decision + migration note (BLOCKED on g2). **→ TODO_LIST Blocked row — M26 overview endgame, owner question g2**
45. After go-nix-helpers FOD fix lands: simplify flake back onto the **→ ROADMAP cross-repo upstream line — go-nix-helpers FOD fix is owner-gated (push never done unasked)**
    go-standard module path (drop manual buildGoModule).
~~46. CI drift-check job wiring for STACK.md (re-run after any new~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
    larsartmann/* require — plan rule 5).
~~47. Consider `nix flake check` in CI (needs the PAT access-token route).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~48. Templ drift gate in CI (`templ generate` + `git diff --exit-code`).~~ done — shipped/closed by the M18-M25 closeout (13-31 archived) and the g3 lint verdict (0+0)
~~49. Fixture drift gate in CI (`scripts/refresh-fixture.sh` + diff).~~ **NOT-DO — fixture drift gate not wired — the golden test pins the committed fixture bytes and the refresh ritual is deliberate (script header); flip when fixture churn becomes real.**
~~50. Re-audit `.golangci.yml` after hub reaches 0: revisit dropped linters~~ **NOT-DO — lint holds 0 with the trimmed set since the g3 verdict; config upkeep folded into fleet lint-policy maintenance.**
    (dupl, wsl_v5, varnamelen…) — add back only with evidence.

## g) Questions for the owner (cannot be figured out from here)

1. **(carried from 07-19 §g)** May I push the one-line go-nix-helpers fix **→ ROADMAP cross-repo upstream line — owner-gated push**
   (remove the redundant `cd "$modRoot"` in `autoDepFodAttrs.modBuildPhase` +
   add a modRoot+deps regression consumer)? The §d4 hash-flip investigation
   may surface a SECOND upstream issue in mkPreparedSource — same question
   applies: push or hand you a patch?
2. **CI secrets:** CI cannot fetch private go-daemon without a PAT. Please **→ TODO_LIST Blocked row — FLAKE_PAT secrets, owner-only**
   create fine-grained `FLAKE_PAT` (read-only Contents on the private repos)
   as a secret on GitHub AND the Forgejo equivalent (or tell me Forgejo's
   runner can reach github.com some other way). I can write everything
   around it; I cannot create secrets.
3. **(carried, g2)** Overview endgame: retire the overview dashboard in **→ TODO_LIST Blocked row — M26 overview endgame, owner question g2**
   favor of this hub, or keep both with a documented split? M27's UX and
   M26's migration note block on this.

---

_Report written 11:21 CET; auto-commit daemon will absorb it. Waiting for
instructions._
