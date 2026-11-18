# go-compliance — Session Status Report (M17–M18 arc)

**Date:** 2026-11-18 07:19
**Session:** Resumed from the 06-47 handoff (M10–M20 report); this session executed M20-F093, all of M17, and drove M18 to mid-bootstrap.

## Scope

**Verified this session (fresh command output, not remembered claims):**

- `go test -count=1 -timeout 300s ./internal/cli/...` → **ok** (36 specs + 2 new bind specs initially red, then green after fixes; final run 0.913s)
- `go test -count=1 ./...` (core module, root) → **ok** after dprint testdata exclusion (golden fixture untouched)
- `dprint check` → **exit 0** after `dprint fmt` normalization (3 files formatted, `testdata/**` excluded from the gate)
- `nix flake lock` → **ok** (go-nix-helpers moved to `github:` URL per fleet truth; `go-daemon` private input pinned to commit `e8e1a1f`)
- `nix build .#compliance-hub` → **FAILING** (mid-bootstrap, see §d — expected state of an F044-style vendorHash loop, not a surprise)

**NOT verified (and why):**

- Full hub suite re-run (`./internal/...`): last full green run was the previous session; this session only re-ran `internal/cli` after editing it. Dashboard/events/cache/discovery suites untouched by this session's edits — green at resume per handoff, not re-proven here.
- `go test -race` over hub: not run (standing debt, TODO_LIST row exists).
- `nix flake check`: blocked on the hub vendorHash bootstrap completing (§b/§d).
- golangci-lint: no config exists yet (debt row exists).
- CI workflows (M19): not started, nothing to verify.

---

## a) FULLY DONE

| Item                                     | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **M20 / F093 — bind posture tests**      | `hub/internal/cli/bind_internal_test.go` (new): `isLoopbackBind` table (9 cases incl. `[::1]`, `0.0.0.0`, `:8477`, `unix:`), unix-socket bind (0o660 perms, dir creation, stale-socket replacement, occupied-dir error), `serveListener` ctx-cancel socket cleanup. `app_test.go` extended: non-localhost warning to errW (incl. decision-doc pointer), loopback default stays silent (Consistently), full `serve --addr unix:...` e2e through the real CLI with socket-removal assertion. Fixed-port 18477 debt killed via `freePort()` helper (all serve specs now dynamic). Suite green. |
| **M17 / F077 — dprint gate**             | `dprint.json` excludes `testdata/**` (golden fixture bytes must not be reformatted — check was red on the fixture, would have silently broken the byte-exact golden test); `dprint check` exit 0.                                                                                                                                                                                                                                                                                                                                                                                           |
| **M17 / F078 — TODO_LIST.md**            | Seeded from the plan + 06-47 report §f debts; open items only, evidence column, BLOCKED section for owner-dependent rows (M26/g2, Forgejo visibility).                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **M17 / F079 — FEATURES.md**             | Full inventory, evidence-cited (`state.go:108`, `events.go:109`, `bind.go:24`, ...); honest statuses: cache PARTIAL (N+1), discovery PARTIAL (watch missing), systemd unit PARTIAL (not deployed); infra PLANNED rows for M18/M19/M21/M22/M24; federation DEMAND_GATED.                                                                                                                                                                                                                                                                                                                     |
| **M17 / F080 — docs/DOMAIN_LANGUAGE.md** | 23 core terms (State, Verdict, Evidence, Fold, Window, Stale, Never-covered, Source, Record, wire v1, Hub, Row, Row diff, Patch stream, Snapshot/Update/Refresh, Discovery, Cache, Probe, Bind posture, Fleet table), bounded contexts (core vs hub), deprecated term (overview).                                                                                                                                                                                                                                                                                                           |
| **M17 / F081 — CHANGELOG Unreleased**    | Replaced scaffold placeholder ("Initial project structure") with the real shipped-work ledger (core module, hub module, security posture); removed the fake `0.1.0 - 2026-01-01` entry (no such release exists — a scaffold lie).                                                                                                                                                                                                                                                                                                                                                           |

## b) PARTIALLY DONE

| Item                            | State                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | What remains                                                                                                                                                                                                                                                                                                     |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **M18 / F082–F085 — flake.nix** | flake.nix written via `flakeModules.go-standard`: devShell (go_1_27, golangci, templ pinned v0.3.1100 via templGen derivation, dprint, actionlint; GOWORK unset for the two-module workspace), treefmt (nixfmt only — Go formatting stays in its dedicated gate), `deps`/`publicDeps` wired for GOPRIVATE shell injection. Hub binary built **manually** via `mkPreparedSource` + `buildGoModule` (modRoot=hub) because of the go-standard FOD bug (§d). Core module covered by `checks.core-build` (compile + `go test` in sandbox). `nix flake lock` green. | vendorHash bootstrap for BOTH derivations (placeholders `sha256-AAA…`/`sha256-BBB…` in place); **the postPatchExtra hub-replace injection is NOT yet on disk** — the python heredoc that writes it failed on quoting (§d) and the subsequent `nix build` ran the old flake. Then `nix flake check` green (F085). |

## c) NOT STARTED

- **M19 / F086–F089** — CI workflows (Forgejo+GitHub compatible; lint pinned to devShell golangci).
- **M21 / F094–F096** — `examples/reporter`, one-line integration guide, README screenshots.
- **M22 / F097–F098** — fold+derive benchmark at 300 repos, baseline under `benchmarks/`.
- **M23 / F099–F100** — BuildFlow TODO_LIST pointer row, STACK.md errata check.
- **M24 / F101–F102** — v2 design doc + compat policy (v1 emittable forever).
- **M25 / F103–F104** — federation research note (hub ↔ go-health-dashboard).
- **M27 / F107–F109** — watch mode (store-write hook, `queryLatestRunCounts` N+1 fix, discovery `/v1/watch`, debounce/coalesce, E2E evidence-change→browser-event).
- **M26 / F105–F106** — BLOCKED on owner question g2 (overview replace-vs-coexist), untouched by design.
- Quality debts: `-race` pass, golangci config, `waitForSubscriber` dedupe, RelativeTime patch purity, .gitignore templ note, AGENTS.md hub conventions, hub `replace` drop at core v0.1.0 tag, fixture-refresh ritual doc.

## d) TOTALLY FUCKED UP!

1. **go-nix-helpers FOD double-cd bug (upstream, blocks the clean path).** `modules/go-standard.nix`'s `autoDepFodAttrs.modBuildPhase` does `cd "$modRoot"`, but nixpkgs' FOD `modConfigurePhase` ALREADY cd'd there — phases share one shell, so the second cd resolves `hub/hub` and fails. Only bites when `modRoot != "."` AND private deps are set (`usePreparedSource`) — an untested combination. Root cause verified against nixpkgs `pkgs/build-support/go/module.nix` (its default FOD buildPhase deliberately does NOT cd). Worked around by building the hub package manually (Option B from the nix-private-go-repos skill). **The proper fix (drop the redundant cd) belongs in go-nix-helpers; shipping it needs a push I will not do unasked.**
2. **Python heredoc quoting failure (my error, this session).** Embedding a nix `'''…'''` block inside a python `'''…'''` heredoc → `SyntaxError` at the `echo` line; the chained `nix build` then ran against the UNPATCHED flake and predictably re-failed on the go-daemon fetch. The postPatchExtra change (hub sub-module replace injection — `replace github.com/LarsArtmann/go-daemon => ../_local_deps/go-daemon` into `hub/go.mod`, the skill-documented `postPatchExtra` route since mkPreparedSource patches only the root go.mod) is **designed but not applied**. Fix is mechanical: apply via the edit tool instead of a python heredoc.
3. **go-nix-helpers' `subModules` naming trap.** The option sounds like "my repo's sub-modules" but actually covers sub-modules OF private DEPS (auto-discovered under `_local_deps/`). The consumer's own secondary module needs `postPatchExtra`. Cost: one build cycle to discover. Recorded here so it is not re-derived.

## e) WHAT WE SHOULD IMPROVE!

- **Fix the go-nix-helpers FOD bug upstream** (one-line: remove the redundant `cd "$modRoot"` from `autoDepFodAttrs.modBuildPhase`, add a regression consumer with modRoot+deps to its self-hosting test) — then this repo's flake can drop the manual buildGoModule block for the module's own path. Needs owner push approval.
- **Document the modRoot+privateDeps combination** in go-nix-helpers README/templates — every two-module repo with a private dep will hit it.
- **Pin the templ generator, not pkgs.templ** (done in this flake via templGen) — consider upstreaming that helper into go-standard as `templPkgOverride` so consumers stop copying the 12-line derivation.
- **Run `go test -race ./internal/...` (hub)** — the events/dashboard concurrency code has never been race-tested; the debt predates this session and the surface grew.
- **golangci-lint config + per-module clean run** — zero lint coverage right now; CI (M19) will need a pinned version anyway.
- **AGENTS.md for this repo** — the verified library API facts (json/v2, go-sse, go-datastar, templ, cmdguard, go-health, go-daemon shapes) live only in the session handoff; a fresh session will re-derive or invent them (6 red cycles were already paid for that).

## f) Top #25 things to get done next!

1. Apply the postPatchExtra hub-replace injection to flake.nix (edit tool, not heredoc).
2. Bootstrap the hub vendorHash (`nix build .#compliance-hub` → paste `got:` hash → repeat until build).
3. Bootstrap the core vendorHash (`checks.core-build`).
4. `nix flake check` green (F085) — M18 complete.
5. Re-run full hub suite + core suite; then M18 flip in TODO_LIST/FEATURES/CHANGELOG.
6. M19: `.github/workflows/ci.yml` — build+test per module (GOWORK=off per-module jobs), valid on Forgejo Actions + GitHub runners.
7. M19: lint job pinned to the devShell golangci version (write `.golangci.yml` first — overlaps the lint debt).
8. M19: dprint + docs-gates (freshness battery) jobs; actionlint the workflow itself (actionlint already in devShell).
9. M21: `examples/reporter/main.go` (uses `compliance.Record`, compiles standalone, go build -o /dev/null test).
10. M21: docs/guides/integrate.md — the one-line-per-project promise.
11. M21: README screenshots (terminal table + WebUI) — render from fixture.
12. M22: benchmark fold+derive at 300 repos (synthetic + fixture), commit baseline under `benchmarks/`.
13. M23: BuildFlow repo TODO_LIST pointer row + link (cross-repo edit).
14. M23: STACK.md errata pass (hosting rows now real).
15. M24: `docs/decisions/v2-wire-design.md` (diff vs v1, compat policy: v1 emittable forever).
16. M25: `docs/notes/federation-interop.md` (research note, hub ↔ go-health-dashboard API surface).
17. M27: fix `queryLatestRunCounts` N+1 (single window query, `hub/internal/cache/cache.go:221`).
18. M27: store write-hook → `hub.PublishRefresh` re-derive.
19. M27: `discovery.Client.Watch` (`/v1/watch` SSE) → debounced refresh trigger.
20. M27: E2E test — evidence change → browser event (datastartest).
21. Quality: `go test -race ./internal/...` hub-wide.
22. Quality: shared `waitForSubscriber` helper; RelativeTime patch-purity check.
23. Quality: AGENTS.md with hub conventions + the verified API facts.
24. Quality: .gitignore note (generated `*_templ.go` committed), fixture-refresh ritual doc.
25. Upstream (owner-gated): fix + push go-nix-helpers FOD double-cd; then simplify this flake back onto the module path.

## g) Top #1 question I can NOT figure out myself!

**May I push the go-nix-helpers fix?** The one-line `cd` removal in `autoDepFodAttrs.modBuildPhase` unblocks the canonical `go-standard` + `modRoot` + private-deps path for this repo (and every future two-module consumer). I will not push unasked; without it, go-compliance carries a manual buildGoModule workaround and the upstream bug stays live for the fleet. (Secondary, still open from last report: owner question **g2** — overview replace-vs-coexist — which gates M26.)

---

_Harvest note: §f items 1–5 are in active execution order; the TODO_LIST.md rows for M17/M20 should be removed as DONE (they live in CHANGELOG now) — next docs-health pass picks that up._
