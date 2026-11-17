# go-compliance

Fleet compliance evidence for every LarsArtmann project: one state model
(`compliant` / `failing` / `stale` / `never-covered`), one frozen wire
contract, a realtime WebUI, and adapters instead of per-project plumbing.

> **Status: pre-inception.** Only the stack decision exists so far — see
> [STACK.md](STACK.md) for the chosen foundation and wiring order. The state
> model, evidence fold, and v1 JSON contract port over from BuildFlow's
> `compliance` command (stays byte-compatible with the overview dashboard).

## Planned shape

- **Core module** (`github.com/larsartmann/go-compliance`) — domain types,
  state derivation over evidence windows, the frozen v1 contract, a `Source`
  adapter interface, and a one-liner reporter SDK. Two dependencies
  (go-branded-id, go-error-family).
- **Hub module** (`github.com/larsartmann/go-compliance/hub`) — HTTP service
  with a realtime WebUI (Datastar over SSE), the CLI (`compliance
  serve|report|table`), a SQLite evidence cache, and adapters: BuildFlow
  evidence (v1 JSON), project-discovery-daemon (fleet list → never-covered).

Integration goal: BuildFlow-covered projects are compliant by default (zero
changes); everything else reports with one CLI line or one SDK call.

## Embedding the dashboard

Any Go service can mount the realtime fleet dashboard in one call:

```go
hub := events.NewHub()                       // publish snapshots/refreshes here
board := dashboard.New(hub)

mux := http.NewServeMux()
board.RegisterRoutes(mux)                    // / (page), /events (patches), /datastar.js
```

Feed it evidence and everything downstream updates live:

```go
rows := events.RowsFromEvidence(evidence, time.Now(), staleAfter)
published, err := hub.PublishRefresh(rows)   // snapshot first, diffs after
```

`RegisterRoutesAt(dashboard.Routes{...}, mux)` relocates the surface (e.g.
behind `/compliance/`); the page route serves JSON to `Accept:
application/json` clients.

Or run the packaged service: `compliance-hub serve --source file --file
compliance.json` (also `--source buildflow`, `--source cache --db path`).

## Documentation

- [STACK.md](STACK.md) — foundation decision record (enforced by
  stack-drift-check)
