# Status: Post-Audit Pareto Plan — Waves 1–3 Execution (Session 2)

**Date:** 2026-11-19 14:43 CET · **Input:** user directive "execute the WHOLE
plan" against `docs/planning/2026-11-19_07-55_post-audit-pareto-execution-plan.md`.

**Scope line (what this report VERIFIED, and what it did NOT):** VERIFIED
live this session: per-task `GOWORK=off go build/vet` + `golangci-lint`
(0 findings) + `go test -race -count=1` for BOTH modules after every
package; `govulncheck ./...` clean on both modules; a FULL `buildflow` run
(findings enumerated below); `actionlint` on the edited ci.yml (exit 0);
`nix flake check` **green at the Wave-2 boundary only**; `templ fmt -fail`
(currently exit 1 — reported in §d). NOT verified: `nix flake check` against
the CURRENT tree (go.mod and root-module Go files changed after the Wave-2
check — the flake is stale until the vendorHash ritual reruns, see §d);
`docs-gates`; `art-dupl` baseline after the file splits; `dprint` on files
edited after the buildflow pass; anything requiring the FLAKE_PAT secret or
a live discovery daemon; browser rendering of the dashboard.

---

## a) FULLY DONE

### Wave 1 — truth & safety (gate green: build+vet+lint+tests, both modules, `-race`)

1. **T1 (1.1–1.3) Discovery SSE parser → `daemon.ParseSSEData`**
   (`hub/internal/discovery/client.go`): WHATWG framing — multi-line
   `data:` frames reassemble (spec'd), over-cap lines (>1 MiB) fail loudly
   with a logged `discovery watch stream failed` (spec'd via a 1.5 MiB
   frame), read errors no longer masquerade as clean EOF; the gopls
   `scannererr` warning is gone (LSP-verified).
2. **T2 (2.1–2.5) Socket-safety, probe-first**: repro-first — wrote the
   second-bind-vs-live-socket test, watched it FAIL (today's steal),
   implemented `removeUnownedSocket` + `daemon.IsSocketAlive` probe on
   `/healthz`, refusal error is a static `errSocketOwned` (errors.Is-able);
   sd_notify/watchdog ordering untouched (`serveListener` unmodified);
   decision recorded (STACK.md: minimal-probe adopted, full `Server.Start`
   declined with reason); go-daemon.md addendum; AGENTS.md updated.
3. **T4 (4.1–4.2) Exit-code contract**: `main.go` now maps through
   `errorfamily.ExitCode` (cmdguard's generic-1 replaced); new suite pins
   the family→sysexits table (IO→75 EX_TEMPFAIL, validation→1,
   internal→69), bare invocation = help = 0, and an e2e
   `report --db /nonexistent` → 75; README gained an Exit-codes section.

### Wave 2 — one deps window (local gate green; `nix flake check` PASSED at the boundary)

4. **T5 (5.1–5.6) Version window**: go-sse v0.6.2→v0.7.0; go-datastar
   v0.6.2→v0.7.0 + datastartest v0.7.0 + static v0.6.1→v0.8.0 lockstep
   (ssetest stays v0.5.0 — its newest tag; go-sse v0.7.0 publishes no
   ssetest). Hand-rolled `/datastar.js` handler replaced by
   `datastar.ScriptHandler()` (ETag/304/HEAD/nosniff — pinned by
   routes_test incl. a 304 revalidation); page script src is versioned
   (`?v=` via `datastar.Version()`), killing the 24h stale-JS window;
   `ringStore` DELETED → `datastar.MemoryStore` (events/store.go and its
   internal test removed, replay suite green); hub vendorHash bootstrapped
   at the package site (converged first paste).
5. **T3 (3.1–3.7) go-filewatcher/v2 v2.5.0 ADOPTED** (public `/v2` module —
   no FLAKE_PAT impact): `cli/watcher.go` dir-watches `<db>`,`-wal`,`-shm`
   (and the evidence file's dir for `--source file` — atomic renames land
   there); cross-process SQLite writes now refresh in ~1–2s instead of the
   5-minute safety net (which STAYS as backstop); `evidence-watch`
   non-critical health check via `watcher.Healthy`; pinned by
   `watcher_internal_test.go`: a write through a SECOND independent
   `cache.Open` connection (invisible to OnWrite) refreshes the hub within
   the 2s budget under `-race`; STACK.md foundation row + decision-log
   entry; go-filewatcher.md addendum; AGENTS.md limitation paragraph
   rewritten (the old "documented limitation" is gone).

### Wave 3 — security & CLI value (verified individually)

6. **T6.1 GO-2026-6859 closed**: `toolchain go1.27.4` directives in BOTH
   go.mod files (verified safe under nix's GOTOOLCHAIN=local: the toolchain
   line is ignored there, local 1.27.3 keeps building; local/CI builds
   auto-select 1.27.4). `govulncheck ./...`: **"No vulnerabilities found"**
   on both modules.
7. **T6.2 SECURITY.md** written (fleet house shape; "Security Practices"
   section named to the checker's schema).
8. **T6.3 nosniff**: `securityHeaders` middleware wraps the whole mux; the
   serve E2E asserts `X-Content-Type-Options: nosniff` on every route.
9. **T7 (7.1–7.4) CLI value**: `table --format` with 11 formats
   (table,json,jsonl,yaml,toml,csv,tsv,markdown,html,xml,asciidoc) via the
   imported go-output sub-module registry; machine formats carry full
   untruncated repo names, RFC 3339 timestamps, zero ANSI; specs cover
   csv/json round-trip/markdown/tsv+html ANSI-free/unsupported rejection.
   7.4 investigated and decided: cmdguard's `--output` is CLI-global and
   NOT visible to command handlers (contract limit) — `--format` stays the
   command-local payload twin; error envelopes already flow through
   cmdguard's format machinery (decision documented in table.go).
   7.3 `health.Off` for cache-disk without `--db`: the check renders
   `status:"off"` with its enable recipe, never gates readiness (readiness
   contract updated to five checks; transition-log spec updated).
10. **file-size findings fixed** (3 error findings): cache.go 376→169+219
    (`queries.go`), events.go 359→302+62 (`handler.go`), reporter.go
    425→217+215 (`filestore.go`) — pure package-internal extractions, all
    gates green after.
11. **cue-vet finding fixed**: `permissions: contents: read` added to
    ci.yml per cue-policy — re-run clean.

## b) PARTIALLY DONE

- **T6.4 buildflow re-run**: done. The THREE target recurring findings are
  confirmed DEAD (govulncheck empty; securitymd reduced to one WARNING
  about response-time wording; the old nosniff finding no longer exists).
  The findings GATE still trips on the remainder: 6× go-humanize
  (**owner-policy-gated**, see §g), 2× docs-gates (routed to T13),
  1× go-structure-linter "-race not configured" (fixed in ci.yml, not yet
  re-verified by a buildflow re-run).
- **T11**: 11.1 CI half done (`-race` added to core+hub test steps);
  AGENTS.md commands row still missing. 11.2 templ-fmt CI job added and
  actionlint-clean — but the NEW GATE IS RED right now (§d). 11.3 (README
  threshold pin) and 11.4 (serve.go comment) not started.
- **Nix state**: `nix flake check` passed at the Wave-2 boundary, but T7's
  import-graph changes (go-output sub-modules direct), the toolchain
  directives, and the core file splits (root module is vendored into the
  hub FOD) came AFTER — **the committed vendorHash is stale for the
  current tree**; the paste ritual must rerun at both sites (T14 window).

## c) NOT STARTED

T10 (citation-check script + run, index screenshot, go-daemon tag-watch
line), T12 remainder (12.2 README link, 12.3 exit-code policy line, 12.4
devShell-build-in-ritual; 12.1 flip-triggers→ROADMAP was already done in
the planning session), T8 (cmdguard-lint CI gate, `AssertDescribeContract`
availability check, upstream Nd TODO — blocked on OG.5), T9 (errorpage,
ThemeToggle, Play CDN→compiled CSS + CSP verification, view.templ:46
embedlit fix, view_test regression + patch-size check), T15 (go-branded-id
+ go-sqlitestore utilization reports), T18 (lychee verdict on the 4
path-excluded files + semantic audit of the 22 token-matched items), T13
(archive sweep + docs-gates clean), T17 (`--profile-io` anomaly), T14
(final ritual). **T21 and OG items remain owner-blocked** (unchanged).

## d) TOTALLY FUCKED UP (known-broken RIGHT NOW)

1. **The new templ-fmt CI gate would FAIL**: `go tool templ fmt -fail
   ./internal/dashboard/` exits 1 — 1 of 4 templ files is unformatted
   (pre-existing drift, surfaced by the new gate). Fix is mechanical
   (templ fmt + regenerate `_templ.go` + commit both together, per the
   AGENTS.md templ rule) and was NOT applied: the user ordered this
   session to stop at this report.
2. **vendorHash stale at both flake sites** (see §b): `nix flake check`
   would fail on the current tree until the bootstrap ritual reruns.
   Everything LOCAL (build/vet/lint/tests) is green; the nix claim is the
   stale one.
3. In-session mistakes (all caught and fixed, recorded for the record):
   table tests initially asserted per-repo severities the frozen v1 wire
   contract cannot carry (adapter.go says so — wasted a test cycle);
   a ci.yml multiedit ate the `hub:` job header (caught on view, restored);
   the first filestore.go split had wrong line boundaries (syntax errors →
   restored my own uncommitted reporter.go and re-split cleanly); two
   edit-races with the auto-commit daemon (re-read + reapply); one
   `rg -rn` hook strike (third-strike warning — must not recur).

## e) WHAT WE SHOULD IMPROVE

- Load the buildflow skill BEFORE the first manual quality command (I ran
  golangci/tidy per project AGENTS for two waves before loading it at
  T6.4; project AGENTS prescribes the manual ritual, but the skill should
  have been read first for the delegation decision).
- Read wire-contract limits (adapter.go "fields the contract cannot
  carry") BEFORE writing data-assertions in tests.
- Inventory ALL import-graph effects of a feature (the `--format` flag
  needed blank imports of four go-output sub-modules → second vendorHash
  window) BEFORE the first nix ritual, to honor "one go.mod window" more
  strictly.
- Treat `nix flake check` as EXPIRED the moment any go.mod or root-module
  Go file changes; say so explicitly in reports instead of leaning on the
  last green boundary.
- Prefer asserted python replaces over `edit` for files the auto-commit
  daemon touches within seconds (two races this session).

## f) Top things to get done next (ordered; harvest input for TODO_LIST/ROADMAP)

1. `templ fmt` write + regenerate `view_templ.go` + commit both (unblocks
   the new CI gate; bundle T9.5's embedlit hint fix in the same templ pass).
2. vendorHash bootstrap at BOTH flake sites + `nix flake check` (after #1).
3. SECURITY.md: add an explicit response-time sentence (kills the last
   securitymd warning).
4. T11.3: pin README health thresholds in the readiness contract test
   (`refreshFreshnessWarn == 10m`, `discoveryStaleWarn == 10m`).
5. T11.4: serve.go goroutine-window clarifying comment.
6. AGENTS.md rows: exit-code policy line (12.3), devShell build in the
   verification ritual (12.4), core `-race` command row (11.1 remainder).
7. `buildflow -s go-structure-linter` re-run (confirm the -race finding
   cleared by the ci.yml change).
8. art-dupl re-audit after the three file splits; verify the moved
   `art-dupl:accept` directive still sits within 5 lines above its clone
   in queries.go; regenerate the baseline only if new clones are
   deliberate.
9. T10.1–10.3: citation-check script over all 20 reports (+ this
   session's addenda); fix/annotate stale citations.
10. T10.4: index.html screenshot (needs firefox permission) + footer link.
11. T10.5: AGENTS.md go-daemon tag-watch line.
12. T12.2: link `docs/review/libs/` from the README docs section.
13. T8.1: cmdguard-lint CI job (verify rule availability first).
14. T8.2: check whether the next cmdguard tag ships `AssertDescribeContract`.
15. T8.3: upstream Nd-duration-codec TODO (BLOCKED on OG.5).
16. T9.1: `errorpage` for Go-default 404s.
~~17. T9.2: `layout.ThemeToggle`.~~ done at 2026-11-19: ThemeToggle + ThemeScript shipped in the compiled-CSS pass (view.templ; pinned by routes_test "links the compiled stylesheet versioned and stays off the Tailwind Play CDN")
18. T9.3–9.4: Play CDN → compiled CSS pipeline + CSP-safety verification.
19. T9.6: view_test regression + patch-size check (rows stay script-free).
20. T15.1a/b: go-branded-id utilization report.
21. T15.2a/b: go-sqlitestore utilization report.
22. T18.1: lychee verdict on the 4 path-excluded files.
23. T18.2–18.4: semantic audit of the 22 token-matched items (3 batches).
24. T13.1: archive sweep (git mv aged reports + manifest).
25. T13.2: docs-gates re-run clean (kills the 2 docs-gates findings).
26. T17.1: `buildflow --profile-io` run on the io-history anomaly.
27. T17.2: route the anomaly finding (fix or ROADMAP).
28. T14.1: `nix flake check --all-systems` once; record aarch64 status.
29. T14.2: final full gate (build/vet/lint/tests both modules + buildflow,
    findings gate modulo owner-gated rows).
30. Owner: go-humanize + samber/lo policy verdict (§g Q1).
31. Owner: FLAKE_PAT secrets on both runner hosts (OG.1) — hub/flake CI
    jobs still skip.
32. Owner: fleet SSE-parser canonical owner (OG.4, §g Q2).
33. Owner: cross-repo upstream mandate (OG.5, §g Q3).
34. Owner: M26 g2 endgame; repo visibility; VACUUM (buildflow preflight
    re-warned: state DB 1.04 GB); cache-DB delete (OG.3).
35. Owner: T21 go signal (core v0.1.0 release pass).
36. dprint check re-run (files edited after the buildflow pass).
~~37. ssetest pin: revisit when a ssetest tag newer than v0.5.0 ships.~~ resolved: routed to ROADMAP "Conditions watched" (ssetest flip trigger row, 2026-11-19)
38. TODO_LIST harvest: mark the plan rows T1/T2/T4/T5/T3/T7 (+T6 partial)
    per this report.
39. go-daemon report rank-3 opportunity (`NewHTTPClientNoTimeout` for the
    watch stream) stays open — flip trigger unchanged.
40. Watch the cross-process watcher spec on CI for timing flakiness (2s
    budget; loosen only with evidence).

## g) Questions I can NOT answer myself

1. **go-humanize (6 error findings) + samber/lo policy** (re-ask of the
   2026-11-18 23-40 §g2 verdict, now concrete): adopt `go-humanize` as a
   real dependency for CLI relative-time formatting (needs a STACK row and
   the `RelTime` swap in health.go/table.go + 3 core sites), suppress the
   linter pattern set, or accept-with-reason? This is now the single
   largest blocker to a green buildflow findings gate.
2. **Fleet SSE-parser canonical owner** (OG.4, re-ask with new evidence):
   Wave 1 made `daemon.ParseSSEData` this repo's consumer choice (zero new
   deps). Should go-daemon own fleet SSE framing (go-sse/sseparse defers
   to it), or the reverse? Decides where any upstream consolidation work
   is filed.
3. **Upstream-TODO mandate** (OG.5, re-ask): may I file the audited
   upstream TODOs in their repos (cmdguard Nd duration codec; go-datastar
   MarshalSignals determinism; templ-components FilterChips + the
   view.templ embedlit hint), or keep withholding all upstream writes?

— Report written 2026-11-19 14:43 CET. Snapshot only; harvest per
docs-health before acting on §f.
