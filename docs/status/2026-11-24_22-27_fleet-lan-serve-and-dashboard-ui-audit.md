# Fleet Run, LAN Serve, UI Diagnosis, templ-components Audit (2026-11-24 22:27)

Session continuation of `2026-11-24_21-34_full-readiness-status-a-to-g.md` (its g1–g3
questions remain open). This session: owner-facing run instructions for
`library-policy fleet --record`, the compliance-hub dashboard on the LAN, a
rendering-health + design diagnosis of the dashboard, and a templ-components
utilization comparison across both repos. Everything below is verified against
source or live commands; no repo code changed except docs/memory.

## a) Fully done

~~1. **Fleet run unblocked for the owner.** Three failure modes diagnosed from the pasted terminal: (a) `/run/current-system/sw/bin/library-policy` is a stale NixOS system-profile build that predates the `fleet` subcommand — `fleet --record` failed "Unknown flag" and bare `fleet` treated "fleet" as a scan path; (b) the default evidence path `/var/lib/overview/compliance.json` is `overview:overview`-owned → per-repo permission-denied cascade; (c) `sudo go run` breaks private module resolution (root has no `GOPRIVATE`; `git ls-remote` gets "could not read Username"). Working recipe verified against `cmd/library-policy/internal/cli/command_fleet.go:31-41` flags: build fresh, run as lars with `--evidence-file ~/.local/state/overview/compliance.json`.~~ done — Verified 2026-11-28: the cited flags still live at library-policy `cmd/library-policy/internal/cli/command_fleet.go:31-41` (`--root`, `--discovery-file`, `--prune` default true; `addEvidenceFlags` carries `--record`/`--evidence-file`); the gotcha set is recorded at library-policy `AGENTS.md:215` ("NEVER `sudo go run`: root lacks `GOPRIVATE`")
~~2. **Hub serve verified end-to-end (loopback).** `nix run .#compliance-hub -- serve --source file --file ~/.local/state/overview/compliance.json`: page 200 ("Fleet compliance"), `/v1/compliance` 200 → 204 repos (157 failing / 47 compliant; library-policy itself compliant); live-reload confirmed (evidence watch + refresher kicks).~~ done — standing session record; the recipe is still canonical (Verified 2026-11-28): `docs/decisions/bind-posture.md:76` (nix run with `--file ~/.local/state/overview/compliance.json`) and `scripts/deploy-hub.sh` deploy the same serve path
~~3. **LAN exposure live, per `docs/decisions/bind-posture.md`.** Explicit `--addr 192.168.1.150:8477` (the doc-mandated loud warning printed — contract verified in logs); bound to the LAN IP ONLY: loopback refuses (curl 000), LAN IP serves 200. Zero firewall change needed — SystemNix `platforms/nixos/system/networking.nix:45` sets `trustedInterfaces = ["eno1"]`, so LAN traffic bypasses the firewall; 8477 collides with nothing in the SystemNix port registry. `dig @127.0.0.1 evo-x2.home.lan` → `192.168.1.150` (dnsblockd), so `http://evo-x2.home.lan:8477` works from LAN devices too. Server runs as lars (no sudo).~~ done — Verified 2026-11-28: `docs/decisions/bind-posture.md:23` ("Non-localhost binds are explicit and loud"), :84-86 (`trustedInterfaces = ["eno1"]`, dnsblockd hostname) and SystemNix `platforms/nixos/system/networking.nix:45` `trustedInterfaces = [ "eno1" ]`
~~4. **UI rendering proven healthy — not a CSS bug.** Objective DOM↔CSS class-coverage audit of the live page (140 real classes): every applied class is styled, including dark mode (`:where(.dark)` rules), motion-reduce fallbacks, and the variant-composed skip link. The −6576-line CSS regen earlier today was benign output tightening (Tailwind v4 `:where()` consolidation), guarded by the CI freshness gate.~~ done — standing account of the 2026-11-24 audit; the CSS was later regenerated again with the @theme re-skin (`hub/internal/dashboard/app.css:42`) and is still byte-freshness-gated in CI (Verified 2026-11-28)
~~5. **UI ugliness diagnosed to root cause.** Stock defaults everywhere: Tailwind's default system font stack verbatim in `app.min.css`, default gray palette + blue accent, library-default Badge/StatCard/Table styling, four equal-weight stat cards burying the actual story (77% of fleet failing), zero visual identity. Not broken — never designed.~~ done — standing diagnosis, confirmed by its fix: CHANGELOG purge row 2026-11-27 — "the generic look is the default theme shipped unmodified, not hand-rolled markup" — closed by the `@theme` block (`app.css:42`) (Verified 2026-11-28)
~~6. **templ-components utilization audit (owner question: "who doesn't use it well?").** library-policy is NOT a consumer: templ-components is `// indirect` in its go.mod (transitive via go-health-dashboard), zero direct imports, no `.templ` files — the matches in `policies/*.go` are ban-rule STRINGS naming the import path. go-compliance IS the consumer, and a good one at the component layer: `view.templ` uses layout.Base, ThemeScript, ThemeToggle, Container, display.Table, display.PageHeader, display.RelativeTime, datastar.SDKScript ×2, datastar.LiveRegion; `statcards.templ` uses display.StatCard ×4 + display.Grid; `rows.templ` maps compliance.State → typed Badge tones (BadgeNeutral/Error/Warning/ Success, `rows.templ:61-73`); `notfound.templ` uses errorpage.NotFound404. Real underuse found: (a) **ZERO `@theme` overrides** in the hub CSS — the library's sanctioned re-skin lever ("theme by overriding Tailwind variables in `@theme`; never add color props to Go") is unexercised, which IS the root cause of the generic look; (b) pin v1.27.0 is one patch behind the released v1.27.2 (ScriptComponent `</script>` escape guard + toast fix); (c) `chips.templ` hand-rolls filter chips — but JUSTIFIEDLY: it needs Datastar signal-mode instant filtering, while the library only ships link/radio-mode FilterChips/SegmentedControl. Candidate upstream contribution, not a defect.~~ done — standing audit, re-verified 2026-11-28; every named underuse since closed: `@theme` at `app.css:42`, pin v1.27.2 (`hub/go.mod:28-31`), the StateBadge→`display.Badge*` mapping now at `rows.templ:76-86`, chip contribution on templ-components master (`5eaa2696`/`bb6b3197`, TODO_LIST Blocked row)
~~7. **Self-correction (provenance).** My earlier critique implied the table, badges, and theme toggle were hand-rolled (DOM class-spotting). Source verification proved all three are library components; the generic look is the library's DEFAULT THEME shipped unmodified. Corrected here and in the chat answer.~~ done — standing, re-confirmed 2026-11-28: `rows.templ:77-86` maps compliance.State onto library `display.Badge*` tones — the table/badges are library components exactly as the correction claims

   **Addendum (2026-11-27, later history; the item above stays as
   written):** every underuse named in item 6 is closed. The hub CSS now
   carries an `@theme` block (one accent, muted severity ramps, system
   fonts), the pin is v1.27.2, the four stat cards are gone (a headline
   sentence, a state bar, and count chips replaced them), and the
   signal-mode chip is being contributed to templ-components. Evidence:
   `hub/internal/dashboard/app.css`, `CHANGELOG.md` Unreleased,
   `docs/status/2026-11-26_17-47_fleet-governance-rev2-execution-status.md`.

## b) Partially done

~~1. **g2 hub deployment.** Ad-hoc LAN serve works (running since ~22:00 as a session-bound background instance); systemd unit, dedicated user, durable bind, and the `/var/lib/overview` ownership decision are NOT done — owner question (see §g).~~ **Superseded — superseded by docs/status/2026-11-26_17-47_fleet-governance-rev2-execution-status.md §f10 — the hub runs as the hardened user unit `deploy/compliance-hub.user.service`; Verified 2026-11-28: `systemctl --user is-active compliance-hub` → active; the SystemNix move stays a TODO_LIST Blocked row.**
~~2. **Corpus hygiene.** This report + TODO_LIST harvest (rows below) written; `docs-gates run .` re-verified green afterwards (result recorded in the final chat message).~~ done — harvest landed and was re-purged at `77c2c45` (2026-11-27; the 17-47 §f6 marker: docs-gates run . all green); on 2026-11-28 audit-all is red again on 30 open items in ten reports 2026-11-18..25 (TODO_LIST High row) — the 2026-11-28 annotate sweeps clear them
~~3. **UI improvement.** Diagnosis + levers identified (@theme tokens first, then hierarchy/typography); implementation NOT started — direction choice pending (§g q1).~~ done — since: the quiet data-first direction shipped via the @theme re-skin (`app.css:42`; CHANGELOG purge row 2026-11-27); §g q1 closed by shipping (Verified 2026-11-28)
~~4. **Memory updates.** AGENTS.md gotchas recorded for both repos (stale system-profile binary, sudo/GOPRIVATE, evidence path; LAN recipe + templ-components verdict) — done this session.~~ done — Verified 2026-11-28: library-policy `AGENTS.md:215` "Fleet run gotchas (re-verified 2026-11-26)" (stale system-profile binary, NEVER sudo go run/GOPRIVATE, /var/lib/overview 0700); go-compliance `AGENTS.md:131` (LAN unicast bind), :36 (deploy-hub.sh unit), :92 (templ-components regen)
~~5. **Dashboard ownership handover.** Takeover instructions delivered (`pkill -f compliance-hub`, then the same nix run with `--addr`), but the running instance is still mine until the owner takes over or g2 lands.~~ **Superseded — superseded by the unit deployment (docs/status/2026-11-26_17-47_fleet-governance-rev2-execution-status.md §f10); Verified 2026-11-28: `systemctl --user is-active compliance-hub` → active — takeover is a systemctl --user restart.**

## c) Not started

Consumer sweep to core v0.2.0 (g3); pushing go-compliance master (~9 commits,
g1 prior session); the @theme/hierarchy design pass itself; screenshot-based
visual QA; upstream datastar-chip contribution; discovery-daemon 400
investigation; fleet dot-dir skip; Findings-column anomaly reconciliation;
evidence timer; library-policy M15/CVE-literal/severity-fork policy items
(routed previously, untouched here).

## d) Totally fucked up

~~1. **Coverage checker shipped with two bugs** (re.escape mangled hyphens; Datastar ternary literals harvested as class tokens) — first output screamed "147/152 classes missing", a false alarm walked back same-session. The instrument was validated only AFTER it scared me.~~ done — standing historical account of the 2026-11-24 session; the discipline stuck — instruments now gate deploys (`scripts/deploy-hub.sh` refuses a red page; CHANGELOG Fixed rows validate the watch spec before shipping) (Verified 2026-11-28)
~~2. **UI attribution without provenance.** I asserted "hand-rolled table / badges / theme toggle" from DOM shape before checking source; partially wrong, corrected only when the owner's templ-components question forced the audit. DOM shape does not prove provenance.~~ done — standing historical account; corrected same-session in §a7 and re-confirmed 2026-11-28 (`rows.templ:77-86` library badge mapping)
~~3. **Session-bound server as "the" LAN dashboard.** The live instance dies with my session and holds port 8477 (takeover needs pkill) — durable ownership should have been transferred immediately, not offered.~~ **Superseded — superseded by docs/status/2026-11-26_17-47_fleet-governance-rev2-execution-status.md §f10 — the hub runs under `deploy/compliance-hub.user.service` (watchdog, LAN unicast bind); Verified 2026-11-28: `systemctl --user is-active compliance-hub` → active.**
~~4. **Stale-binary footgun left armed.** Diagnosed the outdated `/run/current-system/sw/bin/library-policy` but did not fix the underlying SystemNix staleness (needs an owner sudo rebuild).~~ done — no longer true: Verified 2026-11-28 `/run/current-system/sw/bin/library-policy fleet --help` lists fleet with `--evidence-file`, `--discovery-file` and the `$GO_COMPLIANCE_EVIDENCE_FILE` default — the system-profile build is current; no owner rebuild owed
~~5. **Small waste.** Told the owner to rebuild `/tmp/library-policy` when a current build (21:22) already existed.~~ done — standing historical account of the 2026-11-24 session, self-recorded; no follow-up owed (Verified 2026-11-28)

## e) What we should improve

~~1. Validate measurement instruments against a known-good case BEFORE publishing their numbers.~~ done — standing rule, practiced since: the screenshot harness gates deploys (`scripts/deploy-hub.sh` refuses a red page; CHANGELOG Added row) (Verified 2026-11-28)
~~2. Attribute UI findings to source files, never to rendered DOM shape.~~ done — standing rule, practiced: later reports attribute findings to source files (e.g. the 2026-11-27_11-52 report cites `refresher.go` specs) (Verified 2026-11-28)
~~3. Move long-running services to durable ownership (user terminal or systemd) immediately after verification.~~ done — standing rule, practiced: both services moved to durable ownership the next day (`deploy/compliance-hub.user.service`, `fleet-evidence.timer`; Verified 2026-11-28: systemctl --user is-active → active/active)
~~4. Record gotchas to AGENTS.md at discovery time (done here; make it default).~~ done — standing fact, held: gotchas recorded at discovery time (library-policy `AGENTS.md:215` re-verified 2026-11-26; go-compliance `AGENTS.md:131,137-143`) (Verified 2026-11-28)
~~5. Decide obvious small things autonomously (dot-dir skip is just correct) instead of only listing them.~~ done — standing rule, practiced: the dot-dir skip was implemented, not just listed (library-policy `fleet_record.go:155` skips dotted entries; `fleet --prune` default on) (Verified 2026-11-28)

## f) Next things (session-new unless marked carried)

**Fleet ops:**

~~1. Refresh the stale system-profile `library-policy` binary (SystemNix rebuild or PATH fix) — the `fleet` subcommand is missing from the owner's default PATH binary.~~ done — Verified 2026-11-28 `/run/current-system/sw/bin/library-policy fleet --help` lists fleet with `--evidence-file`/`--discovery-file` — the system-profile build is current; no rebuild owed
~~2. Set `GO_COMPLIANCE_EVIDENCE_FILE` in the fleet environment so `--evidence-file` is not needed per run (until g2 decides the real path).~~ done — the env var is the documented override and the default IS the XDG path, so the flag is unneeded — library-policy `AGENTS.md:213` (`--evidence-file` > `$GO_COMPLIANCE_EVIDENCE_FILE` > `~/.local/state/overview/compliance.json`); README.md:40, docs/guides/integrate.md:45, bind-posture.md:70 (Verified 2026-11-28)
~~3. Take the LAN dashboard out of session-bound ownership: owner terminal or systemd (see 21 and §g).~~ done — Verified 2026-11-28 `systemctl --user is-active compliance-hub` → active (`deploy/compliance-hub.user.service`, hardened unit per CHANGELOG Added row); the SystemNix move stays a TODO_LIST Blocked row
~~4. Persist the 204-repo evidence snapshot into library-policy `docs/reports/` (carried TODO row).~~ done — library-policy `scripts/fleet-snapshot.sh` writes `docs/reports/fleet-compliance.md`; the snapshot on disk was regenerated 2026-11-28 05:11 from the timer run (210 repos, producer 4a1ed641) (Verified 2026-11-28)
~~5. Stamp the fleet producer version via ldflags — evidence reports `"buildflow_version":["dev"]` (carried TODO row).~~ done — the nix-built producer stamps its commit into `buildflow_versions` (library-policy `flake.nix:150,206` ldflags); today's snapshot reads "Producer versions: 4a1ed641…"; only hand-built binaries still say dev (TODO_LIST Blocked single-writer row) (Verified 2026-11-28)
~~6. Fleet discovery should skip dot-directories (`.bf-jscpd-worktree` is scanned as a repo and sits in the table with junk gate findings).~~ done — library-policy `cmd/library-policy/internal/cli/fleet_record.go:155` `strings.HasPrefix(entry.Name(), ".")` skips dot-directories (CHANGELOG purge row: 2026-11-25 dot-dir skip with tests) (Verified 2026-11-28)
~~7. Reconcile `.bf-jscpd-worktree`: state `failing` but Findings column renders "—" (verdict/count mismatch — hub display or producer data).~~ done — both halves — the row is dropped by `fleet --prune` (default on, `command_fleet.go:40`; 0 jscpd rows in today's snapshot) and the "—" cell was the severity column, fixed by the findings count (CHANGELOG purge row 2026-11-27, T07) (Verified 2026-11-28)
~~8. Investigate the discovery daemon's `400 Bad Request` on snapshot (`/run/project-discovery/daemon.sock`), logged before `status=pass`.~~ done — root cause was empty search roots — `serve --discover-paths` sends them (`hub/internal/cli/flags.go:39`); CHANGELOG Added/Fixed rows; the 2026-11-27_11-52 report: journal "discovery watch connected", no 400 (Verified 2026-11-28)
~~9. Wire `--discovery-file` so "Never covered" is populated instead of "—".~~ done — fleet `--discovery-file` (`command_fleet.go:36`, "reported as never covered") and hub `--discover-paths` feeding the never-covered set (`flags.go:37-39`) (Verified 2026-11-28)
~~10. Scheduled `fleet --record` timer so stale-after ever has something to bite.~~ done — `fleet-evidence.timer` installed and active — Verified 2026-11-28 `systemctl --user is-active fleet-evidence.timer` → active; snapshot regenerated 2026-11-28 05:11

**Dashboard / templ-components:**

~~11. `@theme` re-skin pass in `library-custom.css`: one deliberate accent +~~ done — `hub/internal/dashboard/app.css:42` carries the `@theme` block (one accent, muted severity ramps, system fonts; CHANGELOG purge row 2026-11-27) (Verified 2026-11-28)
    muted severity ramp, light/dark parity — the sanctioned fix for the
    generic look.
12. ~~Typography decision: intentional typeface (CSP-safe) + `tabular-nums` for counts.~~ done — shipped: the system font stack is set in the `@theme` block of `app.css` (no web font to load, CSP-safe), and `tabular-nums` sits on the chip counts, the findings cells, and the last-run column
~~13. Hierarchy pass: make the failing ratio the page's story; demote~~ done — the stat cards are gone (zero StatCard in `hub/internal/dashboard/`); `view.templ:47` "The headline and the freshness line carry the story" + count chips and per-filter empty states (CHANGELOG Added row) (Verified 2026-11-28)
    zero-value cards (Stale 0, Never covered —) to inline chips.
14. ~~Density pass on the fleet table using library Table knobs (CellPadding/Flush) for 204 rows.~~ done — shipped: `display.TableCellPaddingCompact` with `px-4 py-1.5` row cells and `LazyRows` on the fleet table (`view.templ`, `rows.templ`)
~~15. Bump templ-components v1.27.0 → v1.27.2 (ScriptComponent `</script>`~~ done — `hub/go.mod:28-31` pins templ-components v1.27.2 (CHANGELOG purge row; CSS artifacts regenerated) (Verified 2026-11-28)
    escape guard; toast fix).
    ~~16. Screenshot-based visual QA of the dashboard (pixels, not DOM).~~ done — shipped: `scripts/screenshot.sh` renders light, dark, just-wide, and narrow views in headless Chromium and fails on contrast, keyboard, layout, and console findings; `scripts/deploy-hub.sh` refuses a red page
    ~~17. Decide design direction: quiet/disciplined vs distinctive (§g q1).~~ done — decided by shipping: the dashboard runs the quiet, data-first direction (the `@theme` re-skin in `app.css`: one accent, muted severity ramps, the system font stack). §g q1 was never answered, so the choice stands unless the owner reopens it
16. Contribute a Datastar signal-mode chip variant upstream to **→ routed to TODO_LIST Blocked (owner decisions) row "Adopt the templ-components signal chip once it is tagged" — the chip sits on templ-components master (`5eaa2696`/`bb6b3197`, 2026-11-27) with no release tag; adoption + chips.templ retirement wait on the tag (Verified 2026-11-28)**
    templ-components (justifies retiring hand-rolled `chips.templ`).
~~17. After the pin bump, evaluate StatusBadge/StatusDot vs the local~~ done — decided 2026-11-26 — `display.SimpleEmptyState` adopted; the local `StateBadge` stays because it encodes the four compliance states (`rows.templ:77-86`; CHANGELOG purge row) (Verified 2026-11-28)
    StateBadge→Badge mapping.
~~18. EmptyState/DataState for never-covered/stale empty views.~~ done — `display.SimpleEmptyState` per filter and for the no-evidence case (`view.templ:141`, `view.templ:151`) (Verified 2026-11-28)

**Hub deployment / CI (owner-decision cluster):**

21. Deploy compliance-hub for real: systemd unit (user, bind **→ routed to TODO_LIST Blocked (owner decisions) row "Move both user units (`compliance-hub.service`, `fleet-evidence.timer`) and the library-policy pin into SystemNix" — the unit IS deployed and running (Verified 2026-11-28: systemctl --user is-active compliance-hub → active); the durable/dedicated-user SystemNix decision is the row**
    192.168.1.150:8477 or unix-socket per bind-posture, evidence path) —
    blocked on §g q2.
~~22. Resolve `/var/lib/overview/compliance.json` ownership (overview-user~~ done — by documenting both tiers (2026-11-27): system deployments keep `/var/lib/overview/compliance.json`, user deployments read `~/.local/state/overview/compliance.json`, `GO_COMPLIANCE_EVIDENCE_FILE` overrides both (README.md:40, docs/guides/integrate.md:45, bind-posture.md:70); the user-level hub reads XDG today (Verified 2026-11-28)
    writer flow vs chown/ACL vs XDG default).
23. g1 auth posture: LAN exposure now exists ad-hoc — decide middleware auth **→ routed to TODO_LIST Blocked (owner decisions) row "LAN-trust sign-off for the always-on hub on `192.168.1.150:8477`" — the threat sketch and recommendation are in `docs/decisions/bind-posture.md` ("Sign-off. Open.", line 141) (Verified 2026-11-28)**
    or document LAN-trust acceptance in bind-posture.md.
~~24. Document the LAN serve recipe (bind-posture appendix or runbook: --addr,~~ done — `docs/decisions/bind-posture.md` carries the LAN recipe (explicit `--addr` at :76, `trustedInterfaces`/dnsblockd at :84-86) plus `scripts/deploy-hub.sh` (CHANGELOG purge row 2026-11-25, Appendix A) (Verified 2026-11-28)
    trustedInterfaces note, dnsblockd hostname).
~~25. Discovery-file wiring for fleet + hub (carried).~~ done — fleet `--discovery-file` (`command_fleet.go:36`) and hub `--discover-paths` (`flags.go:39`) wired; CHANGELOG Added row (Verified 2026-11-28)
26. FLAKE_PAT secret for CI hub/flake jobs (carried, owner-only). **→ routed to TODO_LIST Blocked (owner decisions) row "Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs" — still open there (Verified 2026-11-28)**
~~27. Push go-compliance master (~9 commits) and watch 14/14 CI (carried g1).~~ done — master is pushed — Verified 2026-11-28 `git status`: master...github/master 0 ahead, 0 behind; CI at the tip is 13/14 per the TODO_LIST High row (docs gates red on audit-all over 30 open items — the 2026-11-28 annotate sweeps clear them)
28. docs-gates pin parity: CI pin `e305834` lacks the `rows` verb vs the **→ routed to TODO_LIST Low Impact row "Fleet docs-gates pin bump past `e305834` (adds the `rows` verb)" — still BLOCKED (crush-config/fleet nix bump; in-repo writer aliases cover the gap) (Verified 2026-11-28)**
    local binary (carried).
    ~~29. CI drift-check: css pin rev vs flake.lock (carried).~~ done — same fix as the 21-24 report, item 7: the CSS gate job reads the nixpkgs rev from `flake.lock` at run time (2026-11-28)
29. Benchmark re-record decision (carried, owner call). **→ routed to TODO_LIST Low Impact row "Re-record `benchmarks/baseline.txt` (owner call g2)" — VERIFY half done, only the owner re-record call remains (Verified 2026-11-28)**
30. Forgejo repo visibility decision (carried, owner-only). **→ routed to TODO_LIST Blocked (owner decisions) row "Forgejo repo visibility (lars/go-compliance → public?)" (Verified 2026-11-28)**

**Consumers / policies (carried):**

~~32. Consumer sweep → core v0.2.0: library-policy, md-go-validator, BuildFlow~~ done — all four consumers require core v0.2.0 — Verified 2026-11-28 `hub/go.mod:12`, library-policy `go.mod:15`, md-go-validator `go.mod:8`, BuildFlow `go.mod:40`
    (carried TODO row; §g q3).
~~33. library-policy M15: migrate the ADR-0009 gate remnants to `WhenAnyPresent`~~ done — library-policy `policies/ecosystem_combos.go:9` "Migrated 2026-11-20 from the ADR-0009 mixed-list hack"; `WhenAnyPresent` gates at `ecosystem.go:132` and `banned_libraries_doc.go:190-195` (Verified 2026-11-28)
    (still zero call sites).
~~34. Route the 7 raw CVE literals in `policies/security.go` through~~ done — library-policy `policies/security.go:14-24` declares CVEs as `[]policy.CVE` via `mustCVE` → `policy.NewCVE` (parsed once); no raw literals remain (Verified 2026-11-28)
    `policy.NewCVE`.
35. ~~Severity wire-enum fork: `SeverityInfo` sorts as priority 0 and coerces to Moderate on round-trip (routed previously).~~ done — in library-policy (2026-11-20): an exhaustive severity mapper at the DSL bridge; `SeverityInfo` and `SeverityRecommended` no longer sort as priority 0 and "info" no longer coerces (library-policy CHANGELOG, "Severity sort + round-trip"; commit `5f6168e` of 2026-11-24 keeps info companions at info level)
    ~~36. `companion_checker.go:60` invalid `finding.Finding` literal (routed).~~ done — in library-policy at `8812c22` (2026-11-20): missing-companion findings go through `detection.NewMissingCompanionFinding` with a `Validate()` guard (library-policy CHANGELOG, "Invalid `finding.Finding` literal")
36. go-idempotency Q2 parent exemption + Q3 idemtest release prep (owner, prior session). **→ routed to docs/status/owner-decisions-digest.md, row "go-idempotency Q2 parent exemption + Q3 idemtest release prep"**

**Hygiene:**

~~38. Clean my probe artifacts (`/tmp/dash.html`, `/tmp/app.min.css`,~~ done — Verified 2026-11-28 `ls /tmp/dash.html /tmp/app.min.css /tmp/hub-*.json` → all absent; nothing left to trash
    `/tmp/hub-*.json`); check `/tmp/library-policy` freshness before reuse.
~~39. CHANGELOG rows for any tracked docs this session lands (daemon commits).~~ done — CHANGELOG Unreleased carries rows for the docs this session landed (`docs/notes/live-architecture.md`, bind-posture deployment appendix, TODO_LIST purge rows) (Verified 2026-11-28)
40. After the next `fleet --record`, verify the running dashboard live-reloaded
    (watcher contract under real writes).

## g) Questions I can NOT figure out myself

~~1. **Design direction**: quiet and disciplined (tool-grade, data-first) or~~ done — answered by shipping: the quiet data-first direction is live (the `@theme` re-skin in `app.css:42`; the §f verdict on the struck design-direction item); stands unless the owner reopens it (Verified 2026-11-28)
   distinctive (a recognizable identity)? This sets the character of the
   `@theme` token pass — palette, type, emphasis — before I write a line.
2. **Hub deployment shape**: systemd now? Which user (lars vs dedicated), **→ routed to TODO_LIST Blocked (owner decisions) row "Move both user units (`compliance-hub.service`, `fleet-evidence.timer`) and the library-policy pin into SystemNix" — the user-unit shape was decided by shipping (LAN unicast bind, XDG evidence, unit active 2026-11-28); the SystemNix/dedicated-user move is the open decision**
   which bind (LAN IP `192.168.1.150:8477` vs unix socket per
   bind-posture.md), and does the evidence file move to
   `/var/lib/overview/compliance.json` (overview-user writer) or stay under
   `~/.local/state/overview/`?
~~3. **Push + sweep timing**: push go-compliance master (~9 commits) and sweep~~ done — push verified (master...github/master 0 ahead/0 behind, 2026-11-28) and the sweep landed (v0.2.0 in `hub/go.mod:12`, library-policy `go.mod:15`, md-go-validator `go.mod:8`, BuildFlow `go.mod:40`) (Verified 2026-11-28)
   the consumers (library-policy, md-go-validator, BuildFlow) onto core
   v0.2.0 now, or hold both until a natural trigger?
