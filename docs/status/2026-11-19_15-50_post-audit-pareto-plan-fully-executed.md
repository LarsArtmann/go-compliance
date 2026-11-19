# Post-Audit Pareto Plan — FULL EXECUTION STATUS (Waves 1-4 complete)

**Date:** 2026-11-19 ~15:50 CET · **Session:** resume of the interrupted
Wave-3 session (see 14-43) through the T14 final ritual · **Scope:** the
ENTIRE `docs/planning/2026-11-19_07-55_post-audit-pareto-execution-plan.md`
except the owner-gated OG rows and T21's replace-drop (version-currency
gated, see §f1).

## a) FULLY DONE

1. **templ gate + T9.5 embedlit**: `templ fmt -fail` green (5 files); the
   view.templ:46 embedlit hint adjudicated a FALSE POSITIVE — Go forbids
   elided embedded-field composite literal types (proven with a scratch
   program; the templ LSP suggests invalid Go). Explicit form kept; AGENTS
   records the verdict.
2. **vendorHash rituals ×2**: after the Wave-2-green check the tree moved
   go.mod twice more (T7 go-output imports + toolchain, then T9's
   errorpage dep). Both bootstrapped paste-`got:`-converge; `nix flake
   check` green after each; final gate `nix flake check --all-systems`
   exit 0 (6 checks, x86_64 + aarch64).
3. **SECURITY.md response-time wording**: securitymd finding count 1 → 0
   ("Response time: every report is acknowledged within a week...").
4. **T11 remainder**: README readiness thresholds pinned by contract test
   (refresh-freshness + discovery = 10 min; cache-disk floor hoisted to
   `cacheDiskMinFreeBytes` = 128 MiB); serve.go goroutine-window comment
   (launch-before-fallible-wiring rationale); AGENTS rows for core `-race`,
   devShell build, templ fmt gate, exit-code probe (redirect+echo).
5. **Re-verifies**: go-structure-linter `-race` finding CLEARED (5
   pre-existing warning/info suggestions remain, none gate-tripping);
   art-dupl baseline check 0 new clones throughout.
6. **T10 citation infrastructure**: `scripts/check-citations.sh` (423+
   citations mechanically verified per run; ratchet via
   `docs/review/libs/.citations-baseline`); 29 wave-drift citations
   adjudicated (reporter→filestore, cache→queries, events shift, two
   lib-side shadows, one born-stale) with drift addenda appended to 11
   reports; `citation-gate` CI job. T10.4 screenshot: firefox BLOCKED by
   the session permission hook (owner can run it or mark firefox safe).
7. **T8**: `cmdguard usage gate` CI job (pinned `lint/v0.3.0` shallow-clone
   recipe — the published lint module's local-dev replace directives break
   `go run @version`, recorded as an owner-gated upstream TODO);
   `testutil.AssertDescribeContract` verified present on master but NOT in
   the v5.1.0 pin — adoption deferred to the next tag (TODO_LIST row).
8. **T9 templ-components pass**: errorpage adopted (styled 404 for every
   unknown path — page route owns the root exactly via `/{$}`, serve wires
   the catch-all, X-Robots-Tag noindex); ThemeToggle + ThemeScript dark
   mode; Play CDN + inline `tailwind.config` DELETED, replaced by the
   compiled CSS pipeline (`scripts/gen-library-classes.sh` rebuilds the
   tracked class inventory from go.mod truth; `scripts/regen-ui.sh`
   compiles `app.min.css`; served content-hash-versioned at `/app.css`
   with ETag/304/immutable; `css-freshness` CI gate diffs byte-fresh;
   devShell gained `tailwindcss_4`). CSP regression specs: no Play CDN, no
   tailwind.config, script-free row patches under a 2 KB budget.
9. **T15**: go-branded-id (ADOPTED, 90 — the codec core IS the v1 wire
   type spine) + go-sqlitestore (ADOPTED, 85 — engine core owns
   cross-process cache writes) utilization reports; index.html updated to
   22 reports / 13-of-22 in graph, with the go-filewatcher ADOPTED and
   go-daemon widened verdicts reflected.
10. **T18 semantic audit + lychee**: 7 audit-all unharvested items struck
    with evidence (one mis-strike repaired: the writer's first-match-by-
    number hit a duplicate item number — @substring disambiguation is the
    lesson); lychee clean at 0 findings (one real stale `../STACK.md`
    depth link fixed); the "4 path-excluded files" concern is moot under
    the current config — zero exclusions reported.
11. **T13 archive sweep**: 13-31's last open item routed to the owner
    gate; 13-31 + 15-00 (zero open items) `git mv`-ed to
    `docs/status/archived/` with a README manifest; cadence green at 13
    unarchived (room for this report).
12. **T17 io-history anomaly**: NOT-DO with evidence — `io_metrics` shows
    files_hashed tracks tree churn since the previous run (16 quiet / 43
    post-refactor 11-18 evening / 58-27 today). No ratchet defect.
13. **T14 final gates**: core build+vet+`-race`+lint 0 issues; hub
    build+vet+lint 0 issues+`-race` 6/6 ok+templ-fmt green; art-dupl 0
    new; citations 0 stale; dprint clean; actionlint clean; docs-gates
    5/5 PASS; `nix flake check --all-systems` exit 0.
14. **Docs reconcile**: TODO_LIST struck of 27 completed rows (logged in
    CHANGELOG under Unreleased); ROADMAP gained the ssetest watch row;
    CHANGELOG gained the full wave entries + the md-go-validator first
    consumer row.

## b) PARTIALLY DONE

None — every non-owner task in the plan is either fully done or honestly
routed below.

## c) NOT STARTED

None from the plan's agent-executable set.

## d) TOTALLY FUCKED UP (known-broken RIGHT NOW)

Nothing. (The auto-commit daemon raced several edits mid-session — every
one was re-read and re-applied asserted; the docs-gates mis-strike above
was caught by the gate itself and repaired within the session.)

## e) WHAT WE SHOULD IMPROVE

1. The annotate writer's number-only keys silently strike the FIRST
   matching item number across sections — always disambiguate with
   `@substring` in reports with multiple numbered sections (today's
   mis-strike cost one repair cycle; docs-gates caught it).
2. `go get` of a sub-module by bare tag (`errorpage@v1.9.0`) can
   DOWNGRADE the root module when both share tag numbers — always pin
   lockstep (`@v1.27.0`) and re-check go.mod after.
3. The citation baseline grows one line per adjudicated drift; a future
   session could migrate fully-resolved baseline rows into the reports'
   addenda if the ratchet noise ever matters. Not yet a problem (30 rows).

## f) Top things to get done next (ordered; harvest input for TODO_LIST/ROADMAP)

1. Drop hub `replace` when the NEXT core tag ships (v0.1.0 exists but
   predates the Wave 1-3 core splits — filestore.go, toolchain line);
   TODO_LIST carries the sharpened row.
2. Hub `table` `--sort-by` flag parity (pre-existing TODO row, untouched
   by the plan).
3. cmdguard follow-ups at the next tag: `AssertDescribeContract` adoption
   - the lint-module replace-directive upstream fix (BLOCKED row).
4. Screenshot docs/review/libs/index.html (needs firefox permission or
   owner runs it; TODO row).
5. CI observation window: the new css-freshness + citation-gate +
   cmdguard-usage jobs have zero runs yet — watch the first push.
6. docs/status archive cadence: 13 unarchived; close reports' §f
   remainders item-by-item as sessions touch them.

## g) Questions I can NOT answer myself

1. **(carried, 23-40 §g2 / 14-43 §g Q1)** go-humanize adopt/suppress/accept
   — gates 6 error-severity buildflow findings; plus samber/lo policy.
2. **(carried Q2)** fleet SSE-parser canonical owner: `daemon.ParseSSEData`
   vs `go-sse/sseparse`.
3. **(carried Q3)** upstream-TODO mandate: cmdguard Nd codec, go-datastar
   MarshalSignals determinism, templ-components FilterChips — now ALSO the
   cmdguard/lint published-go.mod replace-directive bug (breaks
   `go run @version`; CI works around via shallow clone).
4. **(new)** firefox is permission-blocked in agent sessions — mark it
   safe, or take the index screenshot yourself (T10.4).
5. **(new)** v0.1.0 predates today's core tree — tag v0.1.1 (then the hub
   replace drop executes in ~15m), or direct when.
6. **(carried OG)** FLAKE_PAT secrets; M26 g2 endgame; repo visibility;
   VACUUM + cache-DB deletion; T21 release pass go-signal remains
   satisfied by the existing tag unless you want a re-cut.
