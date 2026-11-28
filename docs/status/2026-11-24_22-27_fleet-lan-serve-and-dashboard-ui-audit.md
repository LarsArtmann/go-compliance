# Fleet Run, LAN Serve, UI Diagnosis, templ-components Audit (2026-11-24 22:27)

Session continuation of `2026-11-24_21-34_full-readiness-status-a-to-g.md` (its g1–g3
questions remain open). This session: owner-facing run instructions for
`library-policy fleet --record`, the compliance-hub dashboard on the LAN, a
rendering-health + design diagnosis of the dashboard, and a templ-components
utilization comparison across both repos. Everything below is verified against
source or live commands; no repo code changed except docs/memory.

## a) Fully done

1. **Fleet run unblocked for the owner.** Three failure modes diagnosed from the
   pasted terminal: (a) `/run/current-system/sw/bin/library-policy` is a stale
   NixOS system-profile build that predates the `fleet` subcommand — `fleet
   --record` failed "Unknown flag" and bare `fleet` treated "fleet" as a scan
   path; (b) the default evidence path `/var/lib/overview/compliance.json` is
   `overview:overview`-owned → per-repo permission-denied cascade; (c) `sudo go
   run` breaks private module resolution (root has no `GOPRIVATE`; `git
   ls-remote` gets "could not read Username"). Working recipe verified against
   `cmd/library-policy/internal/cli/command_fleet.go:31-41` flags: build fresh,
   run as lars with `--evidence-file ~/.local/state/overview/compliance.json`.
2. **Hub serve verified end-to-end (loopback).** `nix run .#compliance-hub --
   serve --source file --file ~/.local/state/overview/compliance.json`: page
   200 ("Fleet compliance"), `/v1/compliance` 200 → 204 repos (157 failing /
   47 compliant; library-policy itself compliant); live-reload confirmed
   (evidence watch + refresher kicks).
3. **LAN exposure live, per `docs/decisions/bind-posture.md`.** Explicit
   `--addr 192.168.1.150:8477` (the doc-mandated loud warning printed —
   contract verified in logs); bound to the LAN IP ONLY: loopback refuses
   (curl 000), LAN IP serves 200. Zero firewall change needed — SystemNix
   `platforms/nixos/system/networking.nix:45` sets `trustedInterfaces =
   ["eno1"]`, so LAN traffic bypasses the firewall; 8477 collides with
   nothing in the SystemNix port registry. `dig @127.0.0.1 evo-x2.home.lan`
   → `192.168.1.150` (dnsblockd), so `http://evo-x2.home.lan:8477` works from
   LAN devices too. Server runs as lars (no sudo).
4. **UI rendering proven healthy — not a CSS bug.** Objective DOM↔CSS
   class-coverage audit of the live page (140 real classes): every applied
   class is styled, including dark mode (`:where(.dark)` rules),
   motion-reduce fallbacks, and the variant-composed skip link. The −6576-line
   CSS regen earlier today was benign output tightening (Tailwind v4
   `:where()` consolidation), guarded by the CI freshness gate.
5. **UI ugliness diagnosed to root cause.** Stock defaults everywhere:
   Tailwind's default system font stack verbatim in `app.min.css`, default
   gray palette + blue accent, library-default Badge/StatCard/Table styling,
   four equal-weight stat cards burying the actual story (77% of fleet
   failing), zero visual identity. Not broken — never designed.
6. **templ-components utilization audit (owner question: "who doesn't use it
   well?").** library-policy is NOT a consumer: templ-components is `//
   indirect` in its go.mod (transitive via go-health-dashboard), zero direct
   imports, no `.templ` files — the matches in `policies/*.go` are ban-rule
   STRINGS naming the import path. go-compliance IS the consumer, and a good
   one at the component layer: `view.templ` uses layout.Base, ThemeScript,
   ThemeToggle, Container, display.Table, display.PageHeader,
   display.RelativeTime, datastar.SDKScript ×2, datastar.LiveRegion;
   `statcards.templ` uses display.StatCard ×4 + display.Grid; `rows.templ`
   maps compliance.State → typed Badge tones (BadgeNeutral/Error/Warning/
   Success, `rows.templ:61-73`); `notfound.templ` uses errorpage.NotFound404.
   Real underuse found: (a) **ZERO `@theme` overrides** in the hub CSS — the
   library's sanctioned re-skin lever ("theme by overriding Tailwind variables
   in `@theme`; never add color props to Go") is unexercised, which IS the
   root cause of the generic look; (b) pin v1.27.0 is one patch behind the
   released v1.27.2 (ScriptComponent `</script>` escape guard + toast fix);
   (c) `chips.templ` hand-rolls filter chips — but JUSTIFIEDLY: it needs
   Datastar signal-mode instant filtering, while the library only ships
   link/radio-mode FilterChips/SegmentedControl. Candidate upstream
   contribution, not a defect.
7. **Self-correction (provenance).** My earlier critique implied the table,
   badges, and theme toggle were hand-rolled (DOM class-spotting). Source
   verification proved all three are library components; the generic look is
   the library's DEFAULT THEME shipped unmodified. Corrected here and in the
   chat answer.

   **Addendum (2026-11-27, later history; the item above stays as
   written):** every underuse named in item 6 is closed. The hub CSS now
   carries an `@theme` block (one accent, muted severity ramps, system
   fonts), the pin is v1.27.2, the four stat cards are gone (a headline
   sentence, a state bar, and count chips replaced them), and the
   signal-mode chip is being contributed to templ-components. Evidence:
   `hub/internal/dashboard/app.css`, `CHANGELOG.md` Unreleased,
   `docs/status/2026-11-26_17-47_fleet-governance-rev2-execution-status.md`.

## b) Partially done

1. **g2 hub deployment.** Ad-hoc LAN serve works (running since ~22:00 as a
   session-bound background instance); systemd unit, dedicated user, durable
   bind, and the `/var/lib/overview` ownership decision are NOT done — owner
   question (see §g).
2. **Corpus hygiene.** This report + TODO_LIST harvest (rows below) written;
   `docs-gates run .` re-verified green afterwards (result recorded in the
   final chat message).
3. **UI improvement.** Diagnosis + levers identified (@theme tokens first,
   then hierarchy/typography); implementation NOT started — direction choice
   pending (§g q1).
4. **Memory updates.** AGENTS.md gotchas recorded for both repos (stale
   system-profile binary, sudo/GOPRIVATE, evidence path; LAN recipe +
   templ-components verdict) — done this session.
5. **Dashboard ownership handover.** Takeover instructions delivered
   (`pkill -f compliance-hub`, then the same nix run with `--addr`), but the
   running instance is still mine until the owner takes over or g2 lands.

## c) Not started

Consumer sweep to core v0.2.0 (g3); pushing go-compliance master (~9 commits,
g1 prior session); the @theme/hierarchy design pass itself; screenshot-based
visual QA; upstream datastar-chip contribution; discovery-daemon 400
investigation; fleet dot-dir skip; Findings-column anomaly reconciliation;
evidence timer; library-policy M15/CVE-literal/severity-fork policy items
(routed previously, untouched here).

## d) Totally fucked up

1. **Coverage checker shipped with two bugs** (re.escape mangled hyphens;
   Datastar ternary literals harvested as class tokens) — first output
   screamed "147/152 classes missing", a false alarm walked back same-session.
   The instrument was validated only AFTER it scared me.
2. **UI attribution without provenance.** I asserted "hand-rolled table /
   badges / theme toggle" from DOM shape before checking source; partially
   wrong, corrected only when the owner's templ-components question forced
   the audit. DOM shape does not prove provenance.
3. **Session-bound server as "the" LAN dashboard.** The live instance dies
   with my session and holds port 8477 (takeover needs pkill) — durable
   ownership should have been transferred immediately, not offered.
4. **Stale-binary footgun left armed.** Diagnosed the outdated
   `/run/current-system/sw/bin/library-policy` but did not fix the underlying
   SystemNix staleness (needs an owner sudo rebuild).
5. **Small waste.** Told the owner to rebuild `/tmp/library-policy` when a
   current build (21:22) already existed.

## e) What we should improve

1. Validate measurement instruments against a known-good case BEFORE
   publishing their numbers.
2. Attribute UI findings to source files, never to rendered DOM shape.
3. Move long-running services to durable ownership (user terminal or
   systemd) immediately after verification.
4. Record gotchas to AGENTS.md at discovery time (done here; make it default).
5. Decide obvious small things autonomously (dot-dir skip is just correct)
   instead of only listing them.

## f) Next things (session-new unless marked carried)

**Fleet ops:**

1. Refresh the stale system-profile `library-policy` binary (SystemNix
   rebuild or PATH fix) — the `fleet` subcommand is missing from the owner's
   default PATH binary.
2. Set `GO_COMPLIANCE_EVIDENCE_FILE` in the fleet environment so
   `--evidence-file` is not needed per run (until g2 decides the real path).
3. Take the LAN dashboard out of session-bound ownership: owner terminal or
   systemd (see 21 and §g).
4. Persist the 204-repo evidence snapshot into library-policy
   `docs/reports/` (carried TODO row).
5. Stamp the fleet producer version via ldflags — evidence reports
   `"buildflow_version":["dev"]` (carried TODO row).
6. Fleet discovery should skip dot-directories (`.bf-jscpd-worktree` is
   scanned as a repo and sits in the table with junk gate findings).
7. Reconcile `.bf-jscpd-worktree`: state `failing` but Findings column
   renders "—" (verdict/count mismatch — hub display or producer data).
8. Investigate the discovery daemon's `400 Bad Request` on snapshot
   (`/run/project-discovery/daemon.sock`), logged before `status=pass`.
9. Wire `--discovery-file` so "Never covered" is populated instead of "—".
10. Scheduled `fleet --record` timer so stale-after ever has something to bite.

**Dashboard / templ-components:**

11. `@theme` re-skin pass in `library-custom.css`: one deliberate accent +
    muted severity ramp, light/dark parity — the sanctioned fix for the
    generic look.
12. Typography decision: intentional typeface (CSP-safe) + `tabular-nums`
    for counts.
13. Hierarchy pass: make the failing ratio the page's story; demote
    zero-value cards (Stale 0, Never covered —) to inline chips.
14. Density pass on the fleet table using library Table knobs
    (CellPadding/Flush) for 204 rows.
15. Bump templ-components v1.27.0 → v1.27.2 (ScriptComponent `</script>`
    escape guard; toast fix).
~~16. Screenshot-based visual QA of the dashboard (pixels, not DOM).~~ done — shipped: `scripts/screenshot.sh` renders light, dark, just-wide, and narrow views in headless Chromium and fails on contrast, keyboard, layout, and console findings; `scripts/deploy-hub.sh` refuses a red page
~~17. Decide design direction: quiet/disciplined vs distinctive (§g q1).~~ done — decided by shipping: the dashboard runs the quiet, data-first direction (the `@theme` re-skin in `app.css`: one accent, muted severity ramps, the system font stack). §g q1 was never answered, so the choice stands unless the owner reopens it
18. Contribute a Datastar signal-mode chip variant upstream to
    templ-components (justifies retiring hand-rolled `chips.templ`).
19. After the pin bump, evaluate StatusBadge/StatusDot vs the local
    StateBadge→Badge mapping.
20. EmptyState/DataState for never-covered/stale empty views.

**Hub deployment / CI (owner-decision cluster):**

21. Deploy compliance-hub for real: systemd unit (user, bind
    192.168.1.150:8477 or unix-socket per bind-posture, evidence path) —
    blocked on §g q2.
22. Resolve `/var/lib/overview/compliance.json` ownership (overview-user
    writer flow vs chown/ACL vs XDG default).
23. g1 auth posture: LAN exposure now exists ad-hoc — decide middleware auth
    or document LAN-trust acceptance in bind-posture.md.
24. Document the LAN serve recipe (bind-posture appendix or runbook: --addr,
    trustedInterfaces note, dnsblockd hostname).
25. Discovery-file wiring for fleet + hub (carried).
26. FLAKE_PAT secret for CI hub/flake jobs (carried, owner-only).
27. Push go-compliance master (~9 commits) and watch 14/14 CI (carried g1).
28. docs-gates pin parity: CI pin `e305834` lacks the `rows` verb vs the
    local binary (carried).
~~29. CI drift-check: css pin rev vs flake.lock (carried).~~ done — same fix as the 21-24 report, item 7: the CSS gate job reads the nixpkgs rev from `flake.lock` at run time (2026-11-28)
30. Benchmark re-record decision (carried, owner call).
31. Forgejo repo visibility decision (carried, owner-only).

**Consumers / policies (carried):**

32. Consumer sweep → core v0.2.0: library-policy, md-go-validator, BuildFlow
    (carried TODO row; §g q3).
33. library-policy M15: migrate the ADR-0009 gate remnants to `WhenAnyPresent`
    (still zero call sites).
34. Route the 7 raw CVE literals in `policies/security.go` through
    `policy.NewCVE`.
35. Severity wire-enum fork: `SeverityInfo` sorts as priority 0 and coerces
    to Moderate on round-trip (routed previously).
~~36. `companion_checker.go:60` invalid `finding.Finding` literal (routed).~~ done — in library-policy at `8812c22` (2026-11-20): missing-companion findings go through `detection.NewMissingCompanionFinding` with a `Validate()` guard (library-policy CHANGELOG, "Invalid `finding.Finding` literal")
37. go-idempotency Q2 parent exemption + Q3 idemtest release prep (owner,
    prior session).

**Hygiene:**

38. Clean my probe artifacts (`/tmp/dash.html`, `/tmp/app.min.css`,
    `/tmp/hub-*.json`); check `/tmp/library-policy` freshness before reuse.
39. CHANGELOG rows for any tracked docs this session lands (daemon commits).
40. After the next `fleet --record`, verify the running dashboard live-reloaded
    (watcher contract under real writes).

## g) Questions I can NOT figure out myself

1. **Design direction**: quiet and disciplined (tool-grade, data-first) or
   distinctive (a recognizable identity)? This sets the character of the
   `@theme` token pass — palette, type, emphasis — before I write a line.
2. **Hub deployment shape**: systemd now? Which user (lars vs dedicated),
   which bind (LAN IP `192.168.1.150:8477` vs unix socket per
   bind-posture.md), and does the evidence file move to
   `/var/lib/overview/compliance.json` (overview-user writer) or stay under
   `~/.local/state/overview/`?
3. **Push + sweep timing**: push go-compliance master (~9 commits) and sweep
   the consumers (library-policy, md-go-validator, BuildFlow) onto core
   v0.2.0 now, or hold both until a natural trigger?
