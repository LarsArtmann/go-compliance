// Package compliance is the fleet compliance core: one state model over
// per-project evidence windows, the frozen v1 wire contract, a Source
// adapter interface, and a one-liner reporter SDK.
//
// The state model ports from BuildFlow's `compliance` command and stays
// byte-compatible with its v1 JSON contract (schema_version 1), which the
// overview dashboard consumes. BuildFlow-schema SQL stays in BuildFlow,
// behind a Source adapter.
//
// Dependency policy: this core module carries go-branded-id and
// go-error-family only. The UI/realtime/CLI stack lives in the hub module
// (github.com/larsartmann/go-compliance/hub). See STACK.md.
package compliance
