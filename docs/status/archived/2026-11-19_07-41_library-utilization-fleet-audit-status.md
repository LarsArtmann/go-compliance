# Library-Utilization Fleet Audit — Status Report

**Date:** 2026-11-19 07:41 CET · **Session:** single sitting, 2026-11-19 ~01:00–07:41
**Scope of this report:** the 20-library utilization audit run ONLY (reports, index,
routing, QC). Not a project-wide health check.

## Scope line — what this report VERIFIED vs did NOT

**Verified by command execution during/after the run:**

- `git fetch --tags` on all 20 library repos (local == origin; tag currency real).
- Consumer dependency picture: root + hub `go.mod`, `go mod why` for sseparse,
  go-atomic-write, go-codec puller chains.
- Independent citation re-verification on **6 of 20 reports**: httputil.md (full),
  go-daemon.md (adoption candidate: `CleanStaleSocket` socket.go:394, `bind.go:46`
  unconditional remove, `ParseSSEData`/`NewHTTPClientNoTimeout` existence), go-sse.md
  (ringStore vs `datastar.MemoryStore`, sseparse-in-graph — one puller citation
  corrected: via go-daemon, NOT ssetest), go-filewatcher.md (should-adopt: OnWrite
  cache-only wiring serve.go:58-60, dir-only contract, WAL `engine.go:216` in the
  pinned module), go-cqrs-lite.md (`DeriveState` state.go:112, appkit TCP-only
  service.go:133-134, refresher ticker comment), go-error-family.md key claim
  (main.go:14 exit path, zero classification consumers in hub).
- index.html: 20/20 links resolve, 6 stat cards, anchors resolve, HTML well-formed
  (parser check), verify block executed line-by-line (the one line I had shipped
  unexecuted — the badge-count rg — was run after the fact: 35).
- Quality gate layers run: `GOWORK=off go build` + `go vet` + `golangci-lint` —
  green on BOTH modules at HEAD (0 findings each).
- TODO_LIST routing: 11 rows containing `docs/review/libs` evidence links confirmed.

**NOT verified (and why):**

- **Visual render of index.html** — `firefox` is permission-blocked in this mode;
  structural/well-formedness/anchor checks ran instead. No in-repo screenshot exists.
- **`go test` suites and `nix flake check`** — not run; every change this session was
  docs-only (.md/.html/STACK.md/TODO_LIST.md/AGENTS.md), reasoned immune. Not proven.
- **Citations in the other 14 reports** — agent self-verified + orchestrator sampling
  per the fleet runbook; not independently re-checked line-by-line.
- Library-side capability inventories were read from checkouts at fetched HEAD; no
  library code was executed.

## a) FULLY DONE

1. **20/20 per-library reports** at `docs/review/libs/*.md`, each with: metadata block
   (reviewed date, group provenance, pins, lib tag + HEAD sha + newest tags after
   fresh fetch), verdict + flip trigger, derived score WITH inputs (M of A, rubric),
   capability inventory (evidence in BOTH repos), gap analysis with Go sketches,
   version currency, fleet-leverage argument (both directions), ranked opportunities,
   executed verify appendix.
2. **Verdicts delivered:** ADOPTED 3 (go-health 94, go-sse 85, go-error-family 75);
   UNDERUSED 5 (templ-components 75, cmdguard 58, go-datastar 58, go-output 57,
   go-daemon 38); NOT-ADOPTED-justified 11 (each with recorded flip trigger);
   NOT-ADOPTED-but-should 1 (go-filewatcher).
3. **index.html dashboard** (Bauhaus Dark kit, marker-based template copy): stat
   cards, 20-row verdict table, cross-cutting findings, method/provenance, verify
   appendix; score-interpretation callout so counterfactual scores can't be misread.
4. **Findings routed at write time:** TODO_LIST gained 4 High/Med + 7 Low rows
   (all evidence-linked); STACK.md stale cqrs-lite module count (70→104) and
   "userless" refresh fixed on sight; AGENTS.md gained the durable audit-outcome
   section (verdict line + re-review triggers).
5. **Fleet-runbook discipline:** 5 sub-agent groups (A–E), concurrency capped at 3,
   no rate-limit failures; QC = first-of-group + every adoption candidate; one
   citation error found and fixed by QC (go-sse sseparse puller).
6. **Cross-group convergence surfaced:** two independent groups (A via go-daemon,
   B via go-sse/sseparse) flagged the same hand-rolled discovery SSE parser
   (client.go:184-201) — promoted to a High-impact TODO row.

## b) PARTIALLY DONE

1. **Phase-8 browser bar** — firefox permission-blocked; fell back to structural
   checks; the in-repo screenshot run-trace does not exist. Fallback one-liner
   documented in my final message, not in the repo.
2. **QC depth** — 6 of 20 sampled (runbook-conformant), but 14 reports' file:line
   citations rest on agent self-verification. One sampled report HAD a factual
   citation error — the base rate on the unsampled 14 is unknown.
3. **Gate** — build/vet/lint green; test suites and `nix flake check` skipped with a
   docs-only rationale (stated, not proven).
4. **Verify-block completeness** — I shipped index.html with one verify line I had
   not yet executed; caught and executed it this morning. Same class of miss the
   skill explicitly warns about.

## c) NOT STARTED (session-derived follow-ons)

~~1. **Upstream routing** — three fleet-wide opportunities identified in reports but~~ done — closed by the post-audit waves (T10 citation script + gate, T15 reports, T12.2 README link, ROADMAP owner-decision routing)
NOT filed in the library repos' TODO lists: cmdguard `Nd` day-suffix duration
codec (36-consumer amortization), go-datastar `MarshalSignals` determinism
option, templ-components signal-mode `FilterChips`.
~~2. **Audit completeness for used-but-unlisted libs** — go-branded-id and~~ done — closed by the post-audit waves (T10 citation script + gate, T15 reports, T12.2 README link, ROADMAP owner-decision routing)
go-sqlitestore are direct deps outside the user's 22-entry list; no reports
exist for them.
~~3. **Mechanical citation check** — no scripted pass verifying every file:line cite~~ done — closed by the post-audit waves (T10 citation script + gate, T15 reports, T12.2 README link, ROADMAP owner-decision routing)
across all 20 reports (the manual QC covered 6).
~~4. **docs/review/libs discoverability** — reachable via AGENTS.md only; no~~ done — closed by the post-audit waves (T10 citation script + gate, T15 reports, T12.2 README link, ROADMAP owner-decision routing)
docs-landing or README pointer.
~~5. **Git-log audit of the run** — the auto-commit daemon committed throughout~~ **NOT-DO — the daemon commits were compile-verified (WP-15) and the tree stayed green throughout.**
(5+ heuristic commits during the run); I never reviewed what it snapshotted
mid-run vs at rest (tree is clean now).

## d) TOTALLY FUCKED UP

Nothing catastrophic — no product code was touched, gate layers run are green.
Honest process stains, worst first:

1. **Shipped a verify block containing an unexecuted command** (index.html badge
   count) — the exact self-refuting pattern the library-deep-dive skill calls out.
   Caught on review the same morning; executed and confirmed (35).
2. **Denominator drift shipped** — index claimed "8 rows into Low Impact"; actual 7.
   Caught by my own re-read during done-bar checks; fixed. Root cause: wrote the
   count from memory instead of counting the table I had just edited.
3. **Browser run-trace bar unmet with a workaround, not a solution** — I treated
   "firefox blocked" as terminal instead of asking the user for a one-time
   permission (the question tool existed).
4. **One real citation error in a sampled report** (go-sse: sseparse puller named
   ssetest; truth: go-daemon) — QC caught it, but it shipped into the report
   initially. Agent claim ("via ssetest") was plausible and wrong.

## e) WHAT WE SHOULD IMPROVE (process, from this run)

1. **Script the citation check.** A 20-line script extracting `file:line` cites from
   reports and verifying existence/range would turn 6/20 sampling into 20/20
   mechanical coverage. Biggest single lever for the next fleet run.
2. **Pre-register count math.** When a report claims "N rows added" to another file,
   count the file immediately — write-time, not review-time.
3. **Execute the verify block BEFORE the file lands** — assemble appendix, run it,
   then write. I ran appends after writing and shipped one gap.
4. **Front-load permission asks.** If the deliverable is HTML, request browser
   permission at dispatch time, not after the failure.
5. **State gate layers explicitly in the report header** ("build+vet+lint green;
   tests+flake not run because docs-only") — I did state it, but making it a
   header-level contract prevents the done-bar from silently shrinking.
6. **Cross-repo finding routing needs an owner rule.** The three upstream
   opportunities have no home; fleet runs should either write the target repos'
   TODO files (with a mandate) or produce a dedicated upstream-findings file.

## f) Next things (up to 50; ranked, session-derived)

**Already routed to TODO_LIST this session (11):**

~~1. Adopt go-filewatcher — dir-watch evidence sources; ~1-2s freshness vs 5-min~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
safety net; default `--source file` has NO write trigger today (High, 2h).
~~2. Replace discovery SSE hand-roll with `daemon.ParseSSEData` (High, 1h).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~3. Stop unconditional unix-socket removal; adopt probe-first `CleanStaleSocket`~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
or `Server.Start` lifecycle (High, 1-2h).
~~4. Fleet exit-code contract: `errorfamily.ExitCode(err)` in main.go (Med, 10m).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~5. `X-Content-Type-Options: nosniff` wrap (Low, 5m).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~6. Bump go-sse → v0.7.0; go-datastar/static → v0.7.0/v0.8.0 lockstep (Low, 30m).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~7. go-datastar de-dup: `ringStore` → `MemoryStore`; `ScriptHandlerWith` ETag (Low, 1h).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~8. `--format` flag on fleet-table path; 16 formats already in graph (Low, 30m).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~9. cmdguard follow-ups: lint CI gate; `AssertDescribeContract` at next tag (Low, 1h).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~10. `health.Off` for unconfigured cache-disk check (Low, 10m).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~11. templ-components wins: `errorpage`, `layout.ThemeToggle`; drop Play CDN (Low, 2h).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)

**New from this review (not yet routed anywhere):**

12. File cmdguard upstream TODO: `Nd` day-suffix duration codec in `Duration` codec. **→ ROADMAP owner decisions — upstream-TODO mandate / fleet SSE-parser canonical owner**
13. File go-datastar upstream TODO: `MarshalSignals` determinism option. **→ ROADMAP owner decisions — upstream-TODO mandate / fleet SSE-parser canonical owner**
14. File templ-components upstream TODO: signal-mode `FilterChips` (chips.templ is **→ ROADMAP owner decisions — upstream-TODO mandate (templ-components FilterChips)**
    its zero-JS design re-implemented).
    ~~15. Write the citation-check script (docs/review/libs quality gate; reusable across~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    fleet runs).
    ~~16. Run it over the 14 unsampled reports; annotate + fix any stale cites.~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
15. Screenshot index.html (needs firefox permission) into docs/assets; link from index. **→ TODO_LIST Low row — screenshot needs firefox permission**
    ~~18. Add docs/review/libs pointer to README or docs landing (discoverability).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    ~~19. Utilization reports for go-branded-id and go-sqlitestore (complete the graph).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    ~~20. Consider `errorpage` replacing Go-default 404s (overlaps #11; split if desired).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    ~~21. After go-filewatcher lands: re-run go-health report row (Healthy() wiring) and~~ **NOT-DO — the wiring is recorded in AGENTS Live Architecture + FEATURES; the go-health report scores the library, not our wiring.**
    update AGENTS.md refresher paragraph (the "documented limitation" shrinks).
    ~~22. After version bumps (#6): regenerate `.art-dupl-baseline.json` only if the~~ **NOT-DO — moot — the bumps and the de-dup shipped in the same go.mod window.**
    de-dup (#7) is NOT done first (order matters; both touch dashboard code).
    ~~23. Re-verify go-daemon report at next go-daemon tag (pin is ahead of v0.4.0; the~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    tag cut should trigger a one-paragraph addendum, not a re-audit).
    ~~24. Fold "flip triggers table" from the 11 justified-no reports into ROADMAP.md as~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
    conditions-watched (they are currently entombed in per-lib reports).
16. Ask: make `docs/review/libs/index.html` a standing artifact refreshed on pin **→ ROADMAP KEEP RAW — calibration family**
    bumps, or point-in-time (current assumption: point-in-time, re-review triggers
    in AGENTS.md).

**Pre-existing TODO_LIST items that this audit makes more urgent (context, not re-routed):**

~~26. GO-2026-6859 toolchain bump + govulncheck re-run (net/http surface audited today).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~27. SECURITY.md creation (bind posture references it via httputil report flip trigger).~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
~~28. Bind-posture open question g1 — it is now the flip trigger for FOUR libs~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
(httputil, go-health-dashboard, go-credentials-store, cqrs-htmx); resolving it
unlocks or permanently closes four adoptions.

**Smaller hygiene noticed in passing (not researched, in-session observations):**

~~29. `view.templ:46` templ hint: unnecessary embedded field type in struct literal~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
(gopls embedlit, displayed in diagnostics all session; trivial).
~~30. `discovery/client.go:184` gopls warning (scanner.Err) — already covered by #2;~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
listed separately so the lint noise disappears with the same fix.
~~31. The `data-tc-relative`/RelativeTime contract is regression-pinned (good) —~~ done — executed by the post-audit waves (T1-T9, T12, T15, T18; 14-43/15-50/16-12 reports)
consider linking that test from templ-components.md's "hardest pattern done
right" paragraph for future readers.
~~32. docs/status archive sweep (pre-existing TODO) now has one MORE report to sweep.~~ resolved: archive rule + manifest landed 2026-11-19 (docs/status/archived/README.md); this report archives once its own §f remainder closes

**Fleet-level (from the audit's cross-repo view; for ROADMAP fuel):**

33. Fleet SSE-parsing canonicalization decision (daemon.ParseSSEData vs sseparse as **→ ROADMAP owner decisions — upstream-TODO mandate / fleet SSE-parser canonical owner**
    direct dep) — affects every project consuming `/v1/watch`-style streams.
34. cmdguard `AssertDescribeContract` adoption wave across CLI consumers at next tag. **→ ROADMAP fleet waves — cmdguard AssertDescribeContract at next tag**
35. go-output `--format` adoption wave (the "formats already in graph" pattern likely **→ ROADMAP fleet waves — go-output --format adoption**
    repeats across cmdguard-based CLIs).
36. go-daemon `Server.Start` adoption wave (socket-steal hazard likely repeats in **→ ROADMAP fleet waves — go-daemon Server.Start adoption**
    every hand-rolled unix-socket service; project-discovery-daemon precedent).
37. go-health `health.Off` pattern for unconfigured deps (fleet-wide idiom). **→ ROADMAP fleet waves — health.Off idiom**

(37 items; honest ceiling for this session's material — padding to 50 would be
invention, not backlog.)

## g) Questions I cannot answer myself

~~1. **go-filewatcher adoption timing:** adopt now (verdict says should; +fsnotify~~ done — adopted same day (T3, dir-watch live)
+3 first-party deps, STACK row, vendorHash ritual at TWO flake sites) or hold
until after core tags v0.1.0 and the hub `replace` drop, so go.mod churn happens
once? Both are defensible; the sequencing is your call.
2. **Fleet SSE-parsing owner:** is `go-daemon.ParseSSEData` or `go-sse/sseparse` **→ ROADMAP owner decisions — upstream-TODO mandate / fleet SSE-parser canonical owner**
intended to be THE canonical client-side SSE parser across the fleet? The hub
can fix its parser either way, but converging every project on one owner is a
fleet decision I can't make.
3. **Cross-repo mandate:** may future sessions write TODO rows into your LIBRARY **→ ROADMAP owner decisions — upstream-TODO mandate / fleet SSE-parser canonical owner**
repos (cmdguard, go-datastar, templ-components) for upstream opportunities like
items 12-14, or do you want upstream findings kept inside consumer repos /
a single fleet backlog you triage?

— End of report. Waiting for instructions.
