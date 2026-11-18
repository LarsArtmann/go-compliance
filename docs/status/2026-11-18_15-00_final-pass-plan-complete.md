# 2026-11-18 ~15:0x — final pass: M27 closed, plan audit clean, all gates green

Scope line: everything below was command-verified this session
(2026-11-18, ~13:35–15:00 CET). The 13-31 report is the predecessor;
this one closes the standing finish-the-plan directive.

## a) FULLY DONE (verified this session)

1. **M27 E2E test verified green** — the uncompiled
   `refresher_internal_test.go` referenced a nonexistent
   `dashboard.RowsID` (real name: `DefaultRowsSelector`) and asserted an
   exact event count of 3 that contradicts the actual stream shape
   (3 initial snapshot + 3 refresh events). Fixed both; the test now
   pins the full contract: initial inner patch carries `boot-repo`,
   live append patch carries `late-repo`. Green under `-race`.
2. **Hub lint back to 0** after refresher/serve additions (one real fix:
   `latestRecordsCapacity` const for the mnd findings in
   `queryLatestRecords`).
3. **`-race` pass over hub: one REAL test-code race found and fixed** —
   `routes_test.go` polled an `httptest.ResponseRecorder` from a gomega
   assertion while the handler goroutine wrote it (the recorder is
   unsynchronized). Replaced with the fleet SDK's own
   `datastartest.CollectWithTimeout` real-request lifecycle.
4. **RelativeTime purity resolved properly** — rows.templ's literal
   leaves `AutoRefresh` false, so row times were FROZEN (no tick script
   existed anywhere). Now: row cells carry the bare `data-tc-relative`
   attribute (script-free patches), and the page ships exactly ONE
   carrier — a "rendered X ago" freshness line whose freezing also
   exposes dead SSE streams. Pinned by a view_test.go regression test.
5. **`waitForSubscriber` deduped** into `hub/internal/testsupport`.
6. **hub `go mod tidy`** — go-daemon + go-health moved to the direct
   require block; the FOD hash proved IDENTICAL (block moves don't
   change the vendored set), so no re-bootstrap was needed.
7. **AGENTS.md + ROADMAP.md created** (hub conventions, FOD
   self-reference lesson, wrapcheck v2 kebab-case gotcha, testing canon;
   adoption triggers, federation build-on-demand options, blocked M26).
8. **Fixture ritual + .gitignore templ note** — refresh-fixture.sh
   header documents the 341-repo ritual (golden bytes, benchmark shape,
   baseline re-record).
9. **docs-health harvest/annotation** — 42 historical open items across
   all six status reports resolved inline with evidence (or routed:
   M26 → TODO_LIST blocked row, adoption meta → ROADMAP); rule-F spacing
   fix in federation-interop.md. **docs-gates battery: 5/5 PASS**
   (freshness, audit-all, annotate, routing, cadence), verified
   formatter-stable (annotate → dprint fmt → re-gate). [annotate count
   reconciliation 2026-11-18 ~22:45: 42 was true at write time; the
   same-day execution pass struck more (13-31 §f2–f6, 20-02 §f4-f8/§g,
   20-02 §25) → docs-gates annotate now reports the live struck count;
   the claim above stays frozen as history]
10. **Plan-vs-tree audit F001–F109** (independent subagent, spot-checked
    artifacts): **106 verified / 3 owner-blocked / 0 missing** —
    F075 (Forgejo visibility), F105/F106 (M26 endgame pair). No gaps,
    no renumbering.
11. **Full closing battery** — root + hub: build, vet, test (hub with
    `-race`), lint 0 issues both; `nix flake check` all checks passed;
    dprint clean.

## b) PARTIALLY DONE

(none)

## c) NOT STARTED

(none — everything actionable is done or owner-blocked)

## d) TOTALLY FUCKED UP

1. One edit-tool slip removed a newline in cache.go (merged two lines);
   caught by immediate re-read and repaired before anything built on it.
2. Two edit races with the auto-commit daemon (mtime churn) — resolved
   via asserted python replaces; no content was lost.
3. My first purity test assertions were wrong three times in a row
   (script string appears twice per script; attribute count includes the
   JS selector literal; the other `<script` tags are the SDK/tailwind
   loads). Debug-rendered the HTML before believing the test — the
   assertions now count what they claim.

## e) WHAT WE SHOULD IMPROVE

1. The exact-event-count assertion pattern in E2E tests is fragile —
   assert contract facts (which patch carries which repo, which mode),
   not raw counts.
2. templ-components could expose an attribute-only "tickable" mode so
   consumers don't hand-roll `Attrs: templ.Attributes{"data-tc-relative": true}`
   (candidate upstream issue).

## f) NEXT

(none — the 27-module plan is fully accounted for; the only open items
live in TODO_LIST: one Low row (drop hub replace when core tags v0.1.0)
and three BLOCKED rows awaiting owner decisions)

## g) OWNER QUESTIONS (unchanged, cannot resolve alone)

1. Push permission for the go-nix-helpers FOD double-cd fix upstream.
2. Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs.
3. M26/g2 overview endgame — replace vs coexist.
4. NEW: local master is 37+ commits ahead of `github/master` (mirror);
   Forgejo origin tracks. Push on your word — standing policy is never
   push unless asked.
