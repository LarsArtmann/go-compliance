# Draft — cmdguard issue: duration codec rejects the `Nd` day suffix

**Repo:** LarsArtmann/cmdguard (own repo) · **Kind:** body-issue · **Not filed** — OG.5 mandate pending

**Proposed title:** `Duration codec can't parse day-suffix values ("14d") — time.ParseDuration passthrough`

## Problem

`ParseDuration` delegates straight to `time.ParseDuration`, which has no day
unit (Go durations cap out at hours):

```go
// skip-validate
// types_duration.go:14 (v5.1.0)
func ParseDuration(s string) (Duration, error) {
	d, err := time.ParseDuration(s)
```

Any fleet-style flag value — `--since 7d`, `--stale-after 14d` — fails the
codec's validation. Consumers end up hand-rolling a day-suffix parser and
bypassing the typed codec.

## Impact

go-compliance's hub carries this exact shim
(`hub/internal/cli/flags.go:120`, `parseDurationSupportingDays`): rewrite
`Nd` → `N*24h`, call `time.ParseDuration`, then feed the result back into
flag defaults the codec should own. Every LarsArtmann CLI with
"days"-shaped knobs repeats this (~10 lines per repo), and each copy
drifts on edge cases (`1d30m` needs a split-parse; `0.5d` nobody agrees on).

## Fix / Proposal

Smallest correct change in `types_duration.go`: pre-expand a trailing
`Nd`/`nd` segment (`7d` → `168h`, `1d12h` → `36h`) before the
`time.ParseDuration` call, keeping every existing value byte-compatible
(`7h`, `168h`, `1h30m` parse identically). Alternative shape if you'd
rather keep the codec stdlib-pure: a `WithFlagCodec`-style option so
consumers register a day-suffix codec per project. The pre-expansion is
the one that deletes code downstream instead of adding an option.
