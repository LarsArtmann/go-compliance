# v2 Wire Design (sketch) + Compatibility Policy

**Status: design sketch — nothing here is implemented yet.**
**Date: 2026-11-18. Owner decision gate: additive evolution only.**

## Context

The v1 document contract (frozen; `v1.go`) is the BuildFlow `compliance
--json` shape overview already parses. It carries per-repo: state, runs,
step rows, failures, fail rate, last-run timestamp/verdict, gate findings,
producer versions, worst steps. It deliberately CANNOT carry: severity
rollups, per-step history, discovery provenance, hub metadata.

Pain points v2 exists to solve (each impossible in v1 without breaking
overview):

1. **Severity data loss** — file-backed writers DROP `FindingSeverities`
   (the v1 wire has no field; the SQLite cache keeps them).
2. **No snapshot identity** — a consumer cannot tell two documents apart
   or request a delta; the hub SSE stream patches are the only delta
   surface.
3. **Flat repo rows** — everything is one flat struct per repo; nested
   per-step evidence (history, timings) has no home.

## Compat policy (the invariant)

**v1 stays emittable forever.** Concretely:

- `MarshalV1`/`DecodeV1Report`/`BuildV1Report` are permanent API. Deprecation
  is forbidden at the type level by policy, not by `// Deprecated:`.
- The default `compliance-hub serve/report/table --source file` paths keep
  producing v1 until overview (the only external v1 consumer) migrates —
  which may be never. v1 is not a legacy format; it is the interop format.
- v2 is ADDITIVE at the document level: a v2 document embeds the v1 object
  (`"v1": {...}`) plus new fields, so a v1-only decoder can read the
  embedded copy while a v2 decoder reads both. No field ever moves or is
  renamed across versions.
- Wire versions are negotiated per surface, never per build: the hub HTTP
  API may offer v2 JSON while the piped document stays v1 indefinitely.

## v2 sketch (hub-facing surfaces only)

The hub module owns v2; the core module stays v1-only (interop surface).
Three shapes:

### 1. v2 snapshot document (`/v2/snapshot.json`, hub)

```jsonc
{
  "schemaVersion": 2, // camelCase: hub JSON is JS-facing (STACK.md)
  "snapshotId": "…", // content hash; delta requests key on this
  "generatedAt": "…",
  "v1": {/* the exact frozen v1 document */},
  "repos": {
    "my-repo": { // map, not array: O(1) patch targets
      "severities": { "error": 1 },
      "importance": "high", // project-meta join (hub cache)
      "lastStateChange": "…", // from the cache's state history
      "steps": { // per-step history (bounded window)
        "lint": { "failRateTrend": [0.5, 0.4] }
      }
    }
  }
}
```

### 2. v2 delta (SSE patch payloads)

The existing Datastar patch stream already IS the delta surface; v2
formalizes: snapshot event carries `snapshotId`; subsequent patch events
reference it; a client with a stale id requests a fresh snapshot (no
partial replay across schema changes).

### 3. Reporter SDK v2 (core-adjacent, hub module)

`RunResult` gains optional severities + step history; the v1 Record path
drops them exactly as today (documented), the cache path stores them.

## What v2 deliberately does NOT do

- No GraphQL/streaming-JSON/schema-registry machinery — single consumer.
- No removal of `schema_version: 1` semantics from v1.
- No second state model — states are the same four; v2 only enriches.

## Implementation order (when the time comes)

1. Hub `/v2/snapshot.json` over the existing cache (read-only, additive).
2. Snapshot identity into the SSE snapshot event.
3. Reporter v2 writes into the cache (severities already stored).
4. overview migration decision — OWNER gate (see M26/g2), independent of
   and never blocking the above.

## Decision ask: reporter-to-cache ingress (T28, 2026-11-27)

**Status: open, owner decision.**

The deployed producer writes the v1 file (`library-policy fleet`, daily
timer) and the hub reads that file. Two things can never become real on
that path:

- **History.** The file store replaces a repository's row on every run, so
  `runs` is always 1 and `fail_rate` is 0 or 1. Staleness and trends need
  the SQLite cache, which keeps every run.
- **Severities.** The frozen v1 contract has no severity field, so the
  dashboard shows a findings count where a severity breakdown belongs.
  The cache stores them (`evidence_runs.severity_counts`, one row per run).

The designed home for both is step 3 above: a reporter mode that writes
into the cache. Concretely it would mean:

| Change                                                  | Where           | Size |
| ------------------------------------------------------- | --------------- | ---- |
| `fleet --cache <hub.db>` records runs into the SQLite cache | library-policy | M    |
| The hub unit serves `--source cache --db <hub.db>`      | deploy script   | S    |
| The producer timer writes cache and file until v1 readers are gone | unit          | S    |

Costs and risks: the producer and the hub then share one SQLite file
(same user; the hub's evidence-watch check already refreshes on WAL
sidecar writes, but a second writer process is untested today and needs a
concurrency spec first), the hub's boot source changes, and the cache needs
a backup story the JSON file never needed.

**The ask:** approve the cache-ingress work (three rows above), defer it,
or reject it in favor of keeping the v1 file as the only producer path.
Until answered, the dashboard keeps the findings-count fallback.
