// Package compliance is the fleet compliance core: one state model over
// per-project evidence windows, the frozen v1 wire contract, a Source
// adapter interface, and a one-liner reporter SDK.
//
// The state model ports from BuildFlow's `compliance` command and stays
// byte-compatible with its v1 JSON contract (schema_version 1), which the
// overview dashboard consumes. BuildFlow-schema SQL stays in BuildFlow,
// behind a Source adapter.
//
// Core surfaces:
//
//   - State/Verdict and DeriveState: the state model (compliant, failing,
//     stale, plus the never-covered extension for repositories a
//     discovery source knows but evidence does not).
//   - RepoEvidence/FoldEvidence: source-agnostic per-repo evidence and
//     the ensure/union fold that BuildFlow's SQL queries feed.
//   - V1Report/BuildV1Report/MarshalV1/DecodeV1Report: the frozen v1
//     wire contract, byte-faithful to the BuildFlow producer and
//     schema-pinned against the overview consumer (golden-tested in both
//     directions against a real producer fixture).
//   - Source/EvidenceWriter and Record: the adapter interface and the
//     one-liner reporter SDK.
//   - FileSource/CommandSource: evidence adapters for piped v1 documents
//     and CLI-spawned producers (the BuildFlow integration).
//
// Dependency policy: this core module carries go-branded-id and
// go-error-family only. The UI/realtime/CLI stack lives in the hub module
// (github.com/larsartmann/go-compliance/hub). See STACK.md.
package compliance
