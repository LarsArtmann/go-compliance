# Roadmap

> Long-term direction and raw ideas not yet refined into actionable
> TODO_LIST tasks. When an item becomes actionable, it graduates to
> TODO_LIST and loses its row here.

## Adoption triggers (from the 2026-11-17 stack-selection scaffold)

- **Tag consumers when adopted** — first real consumers landed
  2026-11-19: BuildFlow bound the dep-light core at `v0.1.0` (see
  `docs/decisions/producer-bind-posture.md`), and md-go-validator's CLI
  adopted the reporter SDK one-liner the same day. Revisit the module
  split and the v1 contract's importance ranking against actual usage;
  adjust before v2 anything.
- **Hub v2 wire** — `docs/decisions/v2-wire-design.md` sketches v2
  (v1 stays emittable forever). Build only when v1 has been solid in
  production for a full adoption cycle.

## Federation (from docs/notes/federation-interop.md, 2026-11-18)

Default posture is passive (A): go-health-dashboard scrapes our
`/health`-equivalent surface; stale-must-not-page holds. Build further
only on a demand signal:

- **B** — source-adapter in go-health-dashboard reading our JSON
  (gated on that repo's maintainers wanting it).
- **C** — link-level federation (hub-to-hub evidence exchange).
- **D** — library convergence (go-compliance core as the shared
  evidence model inside go-health).

## Deduplication and tooling ideas (2026-11-18 art-dupl audit)

Raw ideas harvested from `docs/status/2026-11-18_17-49_art-dupl-dedup-audit-status.md` §f —
each with an explicit keep-raw/groom verdict and demand-signal trigger
(2026-11-18 grooming pass):

- KEEP RAW — art-dupl calibration (`--show-suppressed` judgment pass,
  `-m hash` complement, clone-token trend tracking). Trigger: run the
  calibration pass when the NEXT dedup audit happens, not before; promote
  to a TODO only if it surfaces a suppressed group worth litigating.
- KEEP RAW — `WorstStepLimit` branded type (invalid caps unrepresentable).
  Reason: v1-adjacent public-API churn outweighs the value while the
  defaulting rule has a single tested home (`worstStepLimitOrDefault`);
  reopens with the core v2 API design or a second caller misusing the cap.
- KEEP RAW — templ-components upstream "tickable" RelativeTime mode
  (attribute-only, so consumers stop hand-rolling
  `templ.Attributes{"data-tc-relative": true}`). Trigger: a THIRD consumer
  hand-rolling the attribute, or the next templ-components release window
  (it is that repo's API, not this one's).
- KEEP RAW, demand-gated (unchanged): hub v2 wire — build on a real
  consumer needing >v1 shape; federation B/C/D options — build on
  go-health-dashboard demand (docs/notes/federation-interop.md); the
  art-dupl HTML "template leak" stays closed as NOT REPRODUCIBLE
  (17-49 §f1 annotation) with the re-open trigger recorded there.
- Acceptance policy DECIDED (2026-11-18 execution pass): BOTH mechanisms,
  layered — per-site `// art-dupl:accept <why>` directives stay the
  in-source record of intentional clones, AND a committed
  `.art-dupl-baseline.json` + `art-dupl -t 1 --type-aware check` CI gate
  (job `dupl-gate`, tool pinned v0.9.1-0.20261017225421-93432fd910fb)
  ratchets out hash-known groups so only NEW clones can ever surface.
  Regenerating the baseline is a deliberate act, never a way to silence
  a review finding.

## Endgame question (owner, g2 — BLOCKED)

M26: does go-compliance replace the overview system or coexist?
Answer shapes the migration note and whether an overview-compat
decoder stays a supported surface. See TODO_LIST blocked row.

## Conditions watched — library flip triggers (2026-11-19 audit)

The 11 NOT-ADOPTED-justified verdicts in `docs/review/libs/` each carry a
flip trigger; adopting before one fires is debt. Summary — full argument in
each report: httputil (bind-posture exposure / v2 API), go-etag (snapshot
identity in v2 wire), go-idempotency (first HTTP write endpoint), go-retry
(first flaky one-shot outbound), go-health-dashboard (expose-g1 or
health-hub retirement), go-credentials-store (g1 auth or outbound authed
call), go-appkit (TCP-only core grows a listener seam), go-cqrs-lite (hub
grows a command side), cqrs-htmx (multi-user posture), go-codec (core dep
policy loosening — currently policy-banned), go-atomic-write (same).

Also watched (outside the lib audit): ssetest — the pin sits at the v0.5.0
tag because go-sse v0.7.0 publishes no ssetest; revisit the pin the day a
tag newer than v0.5.0 ships (14-43 §f37). Cross-repo upstream fix
(owner-gated push): the go-nix-helpers FOD double-`cd` bug
(`autoDepFodAttrs.modBuildPhase` re-cd's into an already-cd'd phase —
bites every `modRoot != "."` + private-deps consumer; the one-line fix
would let this flake drop its manual `buildGoModule` block;
07-19 §d1/§e).

## Fleet adoption waves (raw, from the 2026-11-19 utilization audit)

Each likely repeats across LarsArtmann projects; promote to a fleet
backlog/TODO when a second consumer confirms: go-daemon `Server.Start`
adoption (socket-steal hazard in every hand-rolled unix-socket service);
cmdguard `AssertDescribeContract` at its next tag; go-output `--format`
on cmdguard-based CLIs (formats already in those graphs); go-health
`health.Off` idiom for unconfigured deps; the `// art-dupl:accept`
directive convention carried into sibling repos' next dedup audits
(go-etag pinned-by-count precedent). Owner decisions pending: fleet
SSE-parser canonical owner (`daemon.ParseSSEData` vs `go-sse/sseparse`);
mandate for consumer sessions filing upstream TODOs in library repos
(cmdguard Nd duration codec, cmdguard/lint published-go.mod
replace-directive bug, go-datastar MarshalSignals determinism,
templ-components signal-mode FilterChips, a-h/templ LSP embedlit false
positive); **go-humanize adopt/suppress/accept** and **samber/lo
adopt-or-suppress** policy (together the only thing keeping
`buildflow --build-mode=full` from a green findings gate — 6
error-severity findings + ~25 withheld rewrites); **core v0.1.1 tag
timing** (v0.1.0 predates the Wave 1-3 core splits; the hub `replace`
drop executes ~15m after the tag); firefox permission for agent sessions
(index.html screenshot).

## Dashboard/UI raw ideas (2026-11-19 self-review)

Demand-gated, keep-raw: a per-bundle diff-size budget for the
`app.min.css` + `library-classes.txt` churn on templ-components bumps
(catches accidental explosions the byte-freshness gate cannot); a visual
smoke (screenshot diff) for unstyled-class regressions; an explicit
HEAD-request test for `/app.css` (the handler leans on net/http's
implicit HEAD body suppression today).
