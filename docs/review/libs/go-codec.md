# go-codec

| Field            | Value                                                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19                                                                                                                                                                                                                                                                                                                                                                                              |
| Group            | C (library-utilization audit, HTTP/wire-protocol family)                                                                                                                                                                                                                                                                                                                                                |
| Consumer modules | core — **absent**; hub — **INDIRECT**, pinned `github.com/larsartmann/go-codec v0.3.3 // indirect` (hub/go.mod). **Puller identified: `github.com/LarsArtmann/go-daemon v0.4.1-0.20261103232705-e8e1a1fff502`** — the hub's only private dep (sd_notify/unix-socket, `hub/internal/cli/bind.go:15`). Zero consumer code imports it.                                                                     |
| Library version  | repo tag `v0.3.3` (= the pinned version); last commit `98bb1bc 2026-11-15 "chore: auto-commit 1 changed file(s) (heuristic)"`                                                                                                                                                                                                                                                                           |
| Recent tags      | `v0.3.3`, `v0.3.2`, `v0.3.1`, `v0.3.0`, `v0.2.0`, `v0.1.0` (single module, no sub-modules)                                                                                                                                                                                                                                                                                                              |
| Library identity | **Not a JSON-helper library.** It is a `Codec` interface abstraction (CBOR / JSON / Raw) for event-sourced payloads — the serialization contract at the center of a go-cqrs-lite stack (README architecture diagram: consumed by storage/pebble, event, signing, encryption). CBOR (via `fxamacker/cbor/v2`, which go-codec requires at v2.10.0) is the _recommended_ codec; JSON is the interop codec. |

## What go-codec actually IS (read from README + API)

- `Codec` interface (`Encode/Decode/Encoding`) with `CBORCodec`
  (recommended, canonical RFC 7049, signing-safe), `JSONCodec`, `CBORCompactCodec`
  (RFC 8949 core deterministic + unknown-field rejection), `RawCodec`.
- Supporting surface: `BufferEncoder` (zero-alloc), `EncodePooled`,
  streaming CBOR encoder/decoder, `ObservableCodec` (metrics/hooks),
  `AutoDetect`, `TranscodeToJSON`, `WrapEncode/UnwrapDecode` envelopes,
  `Size`, COSE helpers, base64-JSON, deterministic-codec compile-time
  assertions.
- **Its JSON story (decisive for the policy question):** `JSONCodec` is backed
  by `encoding/json` **v1 by default** — `json_compat_v1.go` carries
  `//go:build !goexperiment.jsonv2`; the `encoding/json/v2` path
  (`json_compat_v2.go:1 //go:build goexperiment.jsonv2`) only compiles under
  the `GOEXPERIMENT=jsonv2` build tag. So in a default build, adopting
  go-codec's JSON path would route the hub's JSON through a wrapper around
  _v1_ stdlib JSON — a regression against both policy and byte-compat, not a
  neutral choice.

## The policy engagement (STACK.md: "JSON: encoding/json/v2 (stdlib), GA in Go 1.27; all third-party JSON libs banned")

Non-adoption in core's frozen v1 contract is **policy-correct, plainly**:
core's `MarshalV1` calls `encoding/json/v2` directly (`v1.go:5`) and its
byte-exactness is pinned by a golden fixture with 341 real repos
(`testdata/buildflow-v1-fixture.json`, `v1_golden_test.go`) plus
structure-order comments that exist _because_ the wire bytes are load-bearing
(`v1.go:29` struct-declaration-order under json/v2; `v1.go:70` map-key
sorting). Any `Codec`-shaped indirection in that path adds a translation
layer between the golden fixture and the stdlib behavior it pins, for zero
capability gain — the hub's other JSON surfaces (SSE event payloads
`hub/internal/events/events.go:5,37`, discovery wire types
`hub/internal/discovery/client.go:11`, cache column marshaling
`hub/internal/cache/cache.go:11`) are likewise direct `encoding/json/v2`.

Strictly, go-codec is first-party (larsartmann), so the "third-party JSON
libs banned" line does not name it; but the policy's _mechanism_ — JSON goes
through stdlib so behavior is exactly what the toolchain documents — applies
with full force, and go-codec's default-v1 JSON path fails even the letter of
the intent. Verdict on direct adoption for JSON: banned-equivalent, correctly
so.

## Is the indirect pin a smell? Whose problem is it?

**Not a smell, and it is go-daemon's problem.** Evidence chain (executed):

```
$ cd hub && GOWORK=off go mod why github.com/larsartmann/go-codec
# github.com/larsartmann/go-codec
github.com/larsartmann/go-compliance/hub/internal/cli
github.com/LarsArtmann/go-daemon
github.com/larsartmann/go-codec

$ GOWORK=off go mod graph | grep go-codec | head
github.com/larsartmann/go-compliance/hub github.com/larsartmann/go-codec@v0.3.3
github.com/LarsArtmann/go-daemon@... go-codec@v0.3.3
...
```

The hub imports go-daemon for `bind.go`'s sd_notify/watchdog surface; go-daemon
requires go-codec (plausibly for its own socket-protocol payloads) which
requires `fxamacker/cbor/v2` (v2.9.6 required, MVS-resolved to v2.10.0 — also
the version go-daemon itself requires). The `// indirect` pin in hub/go.mod is
normal MVS bookkeeping recording the resolved version. It becomes
go-compliance's problem only if: (a) go-codec (or cbor) grows a CVE — then the
fix is a `go get github.com/larsartmann/go-codec@<fixed>` bump, same as any
transitive; (b) the binary-size/attack-surface cost of carrying cbor+float16
into a hub binary ever matters (it does not: pure-Go, small); or (c)
STACK.md's policy were interpreted to ban non-stdlib JSON _anywhere in the
build graph_ — it is not; the policy governs _this repo's chosen JSON path_,
and go-daemon's internal payload encoding is go-daemon's policy domain (the
same reasoning by which the hub already accepts go-daemon's other transitive
deps, e.g. cmdguard → samber/do, recorded as "transitive only" in AGENTS.md).

One honest caveat worth recording: go-daemon's own payloads — if it uses
`JSONCodec` anywhere wire-facing — run on encoding/json **v1** by default.
That is a byte-stability question for go-daemon's socket protocol, not for
go-compliance's wire contract; the hub's wire surfaces never pass through it.

## Non-JSON capabilities the hub could use? (assessed separately)

- **CBOR for the SQLite cache**: `gate_findings`/`severity_counts` columns are
  JSON TEXT today (`marshalCounts`, `hub/internal/cache/queries.go:168-183`).
  CBOR would shrink them and add unknown-field rejection, but requires BLOB
  columns + a migration in the append-only `migrations` list (cache.go:26-48),
  loses `sqlite3` CLI readability, and saves bytes nobody counts at 341 repos.
  N/A by payoff.
- **Deterministic signing / envelopes / COSE / streaming batches / autodetect**:
  no signing, no envelope protocol, no mixed-format streams exist in this
  consumer. N/A.

## Verdict — **NOT-ADOPTED-justified** (core: policy-correct absence; hub: no capability fit; indirect pin: benign, owned by go-daemon)

**Flip triggers:** (a) core ever grows a _second_ wire encoding deliberately
(e.g. a CBOR sidecar for the SSE patch stream to cut browser payload) — the
STACK.md policy row would need a decision-log update first, and go-codec's
CBOR would be the candidate; (b) go-compliance adopts direct ownership of a
payload-encoding contract for embedders (no sign of this); (c) go-daemon ever
drops go-codec — the indirect pin disappears with it, no action needed here.

## Derived score (counterfactual: what adopting would cover vs already solved)

Bands: 85-100 excellent · 70-84 good · 50-69 fair · <50 poor.

| Capability                            | Status                | Library evidence (path)                        | Consumer evidence (path:line)                                                                                                |
| ------------------------------------- | --------------------- | ---------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| JSON encode/decode                    | Covered elsewhere — 1 | `json.go` (`JSONCodec`)                        | direct `encoding/json/v2` at every JSON site: `v1.go:5`, `events/events.go:5`, `discovery/client.go:11`, `cache/cache.go:11` |
| Byte-stable wire marshaling           | Covered elsewhere — 1 | deterministic codecs, golden tests             | golden fixture byte-exactness pinned natively (`v1_golden_test.go`, `testdata/buildflow-v1-fixture.json`, 341 repos)         |
| CBOR payloads                         | N/A                   | `cbor.go` (`CBORCodec`, recommended)           | no CBOR surface exists in the consumer; SQLite cache uses JSON TEXT (`cache/queries.go:168-183`)                             |
| Compact/schema-drift-rejecting decode | N/A                   | `cbor_compact.go`                              | all inbound shapes are frozen contracts with explicit decoders (`DecodeV1Report`, discovery wire structs)                    |
| Zero-alloc / pooled encoding          | N/A                   | `pool.go`, `BufferEncoder`, `streaming.go`     | hot paths are SSE event sends (already go-sse) and a 5-min-cadence derive — no alloc-bound encode path                       |
| Observable codec (hooks/metrics)      | N/A                   | `observability.go`                             | no codec telemetry need                                                                                                      |
| Transcode CBOR↔JSON                   | N/A                   | `transcode.go`                                 | no mixed-format data                                                                                                         |
| Envelope / COSE / signing safety      | N/A                   | `envelope.go`, `cose.go`, `DeterministicCodec` | no payload signing in the consumer (localhost/socket posture, `docs/decisions/bind-posture.md`)                              |

Inputs: N=8, A=2 (both "Covered elsewhere", Σ=2.0) → **score = round(100×2.0/2)
= 100 (excellent)** counterfactual: the only capabilities in go-codec's domain
the consumer needs (JSON, byte-stable marshaling) are already met by the
stdlib directly — which is the policy working as intended. Publish: **2 of 2**.

## Gap analysis

No actionable gaps. Two notes for the record:

1. **Do not "clean up" the indirect pin.** Removing or pinning it manually
   would fight MVS and break the go-daemon build; it rides go-daemon's
   requirements and should drift only via `go get github.com/LarsArtmann/go-daemon@…`
   bumps. (AGENTS.md already notes go-daemon's require sits in the indirect
   block pending tidy — same class of bookkeeping.)
2. **If a CVE ever lands in go-codec/cbor:** bump the transitive pin
   (`cd hub && GOWORK=off go get github.com/larsartmann/go-codec@<fixed>
   && GOWORK=off go mod tidy`), then re-bootstrap the Nix vendorHash at BOTH
   flake sites (AGENTS.md "Nix / FOD Lessons"). That is the entire runbook.

## Version currency

Pinned indirect at v0.3.3 = latest tag exactly; no drift. Library active
(last commit 2026-11-15; tags v0.1.0→v0.3.3 over its history) and heavy on
contract tests (deterministic v1/v2 suites, fuzz, goldens). The pin will move
only when go-daemon moves — nothing for this repo to track.

## Fleet leverage

For: go-codec is the serialization contract for the cqrs-lite half of the
fleet; services in that family get signing-safe determinism and pooled
encoding from one place. Against: go-compliance is deliberately _not_ in that
family — its serialization story is "frozen stdlib JSON bytes, proven by a
golden fixture", the opposite of an abstraction. Pulling the contract layer
in would create a second place where wire behavior is defined. The fleet is
best served here by _not_ consolidating: the pin exists only as cargo on the
go-daemon truck, which is exactly right.

## Ranked opportunities (impact × ease)

1. **Nothing now.** Optional: add one line to STACK.md's decision log noting
   "go-codec rides go-daemon (transitive; policy governs our direct JSON
   path)" if the indirect larsartmann pin ever trips the stack-drift-check —
   as of this audit the check only covers _direct_ requires, so no entry is
   required.

## How to verify (all commands executed 2026-11-19)

```bash
# Library identity & nature
git -C /home/lars/projects/go-codec describe --tags --abbrev=0        # → v0.3.3
git -C /home/lars/projects/go-codec log -1 --format='%h %cs %s'       # → 98bb1bc 2026-11-15 chore: auto-commit...
git -C /home/lars/projects/go-codec tag --sort=-creatordate | head -5 # → v0.3.3 v0.3.2 v0.3.1 v0.3.0 v0.2.0
head -60 /home/lars/projects/go-codec/README.md   # "Payload Encoding (CBOR / JSON / Raw)" — not a JSON helper lib
cat /home/lars/projects/go-codec/go.mod           # requires fxamacker/cbor/v2 v2.10.0 (+ test-only deps)

# The JSON-default-is-v1 claim (the policy-relevant fact)
sed -n 1,3p /home/lars/projects/go-codec/json_compat_v1.go   # → //go:build !goexperiment.jsonv2  (default build)
sed -n 1,3p /home/lars/projects/go-codec/json_compat_v2.go   # → //go:build goexperiment.jsonv2    (opt-in only)

# Indirect pin + puller (the audit's required commands)
cd /home/lars/projects/go-compliance/hub && GOWORK=off go mod why github.com/larsartmann/go-codec
#   → hub/internal/cli → LarsArtmann/go-daemon → go-codec
cd /home/lars/projects/go-compliance/hub && GOWORK=off go mod graph | grep go-codec | head
#   → hub→go-codec@v0.3.3 (indirect bookkeeping); go-daemon@…→go-codec@v0.3.3 (the real require)
cd /home/lars/projects/go-compliance && GOWORK=off go mod why github.com/larsartmann/go-codec
#   → (main module does not need package …) — core is clean
rg -n 'go-codec' /home/lars/projects/go-compliance/hub/go.mod          # → v0.3.3 // indirect
rg -n 'larsartmann/go-codec' /home/lars/projects/go-compliance --type go -g '!vendor/**'  # → no source imports

# Policy + consumer JSON sites
rg -n 'encoding/json' /home/lars/projects/go-compliance/v1.go /home/lars/projects/go-compliance/hub/internal/events/events.go /home/lars/projects/go-compliance/hub/internal/discovery/client.go /home/lars/projects/go-compliance/hub/internal/cache/cache.go
#   → all import "encoding/json/v2" directly
rg -n 'all third-party JSON libs banned' /home/lars/projects/go-compliance/STACK.md  # → the constraining policy line

# go-daemon's role in the hub
rg -n 'go-daemon' /home/lars/projects/go-compliance/hub/internal/cli/bind.go  # → :15 import (sd_notify/watchdog)
```

Spot-checks: hub/go.mod indirect block read (go-codec v0.3.3, cbor v2.10.0);
bind.go:15 import verified; both `go mod why` chains read from executed
output; v1.go:5 / events.go:5 / client.go:11 / cache.go:11 imports read
directly. The MVS-resolved cbor version (v2.10.0 vs go-codec's v2.9.6 floor)
verified from `go mod graph` output showing go-daemon itself requires
v2.10.0.

## Citation drift addendum (2026-11-19 post-audit wave)

The `hub/internal/cache/cache.go` marshalCounts citation (324-338) moved
to `hub/internal/cache/queries.go:168-183` in the cache split (verified
by `scripts/check-citations.sh`).
