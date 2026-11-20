# Draft — cmdguard issue: published lint module ships directory `replace` directives

**Repo:** LarsArtmann/cmdguard (own repo) · **Kind:** body-issue · **Not filed** — OG.5 mandate pending

**Proposed title:** ``lint module publishes `replace => ../` directives — `go run @version` unusable``

## Problem

The published `lint` module's `go.mod` carries four directory replaces
(`lint/go.mod:81-87`, verified at `lint/v0.3.0`):

```go
replace github.com/larsartmann/cmdguard/flightrecorder => ../flightrecorder
replace github.com/larsartmann/cmdguard/glamour        => ../glamour
replace github.com/larsartmann/cmdguard/spinner        => ../spinner
replace github.com/larsartmann/cmdguard/v5             => ../
```

Go does not strip `replace` directives on publish, so the module proxy
serves them verbatim. Every consumer invocation through the proxy then
demands local sibling checkouts that don't exist:

```
go: github.com/larsartmann/cmdguard/lint@v0.3.0 (in github.com/larsartmann/cmdguard/lint@v0.3.0):
	replace github.com/larsartmann/cmdguard/v5 => ../: directory ../ is outside the module or workspace
```

`go install`/`go run ...@version` — the entire point of publishing a lint
tool — is unusable.

## Impact

go-compliance's CI `cmdguard-usage` gate can't run the tool the documented
way; the working invocation is a shallow clone of the pinned tag plus
`go -C /tmp/cmdguard/lint run ./cmd/cmdguard-lint`
(`.github/workflows/ci.yml:175-177`). That works but pins every consumer
to git-clone mechanics and re-clones ~the whole monorepo per CI run.

## Fix / Proposal

Smallest correct change: a release step that strips the sibling replaces
and re-points them at real versions before tagging —
`go mod edit -dropreplace=... -require=...@v5.1.0` for the four lines in
a `lint/release` target (or a GoReleaser hook), then tag from the rewritten
tree. Worth confirming the sibling modules resolve at real versions first;
if one has no usable tag yet, that's the actual release blocker. Until
then, documenting the clone-based invocation in the lint README would make
the workaround official instead of discovered.
