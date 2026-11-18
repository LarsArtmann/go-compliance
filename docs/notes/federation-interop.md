# Federation Interop: compliance hub ↔ go-health-dashboard

**Status: research note (M25/F103-F104). No code commitment.**
**Date: 2026-11-18.**

## What go-health-dashboard federation actually is

Verified against the go-health-dashboard repo (2026-11-18 checkout):

- The **`aggregate` package** (`github.com/larsartmann/go-health/aggregate`)
  merges N local `Prober`s into one surface `dashboard.New` renders
  (`aggregate.Source{Name, Probe}`). It is an IN-PROCESS merge, not a wire
  protocol.
- The **`health-hub` binary** (`cmd/health-hub`) is the cross-machine
  federator: it scrapes each remote service's `/health` (and
  `/health/export`) JSON — the "byte-stable scrapes" determinism exists
  exactly for this — and renders one dashboard. Ingestion is typed to
  go-health's shape (`SourceStatuses`, per-check metadata, `StatusSince`);
  it does not parse arbitrary JSON.
- Federation hubs deliberately drop per-process scalars (instance IDs)
  because they would lie about a merged view — same instinct as our
  snapshot identity work.

## The two domains (why direct interop is not automatic)

|                    | go-health-dashboard          | compliance hub                  |
| ------------------ | ---------------------------- | ------------------------------- |
| Question answered  | "is service X up, right now?" | "did repo Y comply, this window?" |
| Time model         | instant probes, `StatusSince` | evidence windows, staleness      |
| Unit               | service / check              | repository / step                |
| Update cadence     | seconds                      | per-run (minutes to days)        |

A repo going stale-for-14-days is NORMAL compliance signal; a service
silent for 14 days is an incident. Naively mapping states
(failing→unhealthy) would page people for old evidence.

## Interop options

### A. Passive scrape (today, zero work)

The compliance hub already builds a go-health self-probe (`/health`,
serve.go). A fleet health-hub ingests the hub as one remote: operators see
"compliance-hub is up" next to their services. No fleet states cross over.

- Cost: zero. Coupling: go-health probe only. **This is the default.**

### B. Compliance-as-a-source inside health-hub

A `Prober` adapter (living in go-health-dashboard or health-hub, NOT
here) that scrapes the compliance hub's `Accept: application/json` page
route and maps each repo row to a check. Tradeoffs:

- One pane for everything (the actual want behind every federation ask).
- State semantics must be translated with intent: `failing` maps to a
  degraded check ONLY if the last run is fresh; `stale`/`never-covered`
  must map to informational, never unhealthy — otherwise old evidence
  pages on-call.
-	Version skew risk: the compliance JSON surface is v1-frozen, so the
  adapter pins cleanly; v2 (docs/decisions/v2-wire-design.md) is additive.

**Gate: owner decision — nobody has asked for one pane yet.**

### C. Link-level federation (cheap, honest)

Both UIs are templ + templ-components: converge the visual language
(severity chips, statcards) and cross-link (health dashboard links to the
compliance hub and back). Domains stay separate; operators get one
bookmark bar, not one pane.

### D. Library convergence (ongoing, already happening)

Both stacks independently chose go-datastar + go-sse + templ-components +
cmdguard. Component-level reuse (chips, live badges) benefits both without
any wire coupling. The STACK.md exclusion of importing go-health-dashboard
stands: same architecture, different domain.

## Recommendation

A now (it exists), C whenever the visual gap annoys, B only behind an
owner ask with the stale-must-not-page rule written into the adapter, D as
background convergence. Record any B decision in both repos' STACK files.
