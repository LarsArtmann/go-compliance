# Release runbook (go-compliance)

The step-by-step sequence for cutting a release. `scripts/pre-release-check.sh`
encodes the gate order mechanically — this doc explains the WHY behind each
step, distilled from the v0.2.0 run (2026-11-20) and the v0.1.0 tag-only gap.

## Release shape

- **Core module** (repo root): released via annotated `vX.Y.Z` tags, served
  by the module proxy, pkg.go.dev renders it. This is the only go-get-able
  surface.
- **Hub module** (`hub/`): deliberately NEVER tagged (`hub/vX.Y.Z` tags are
  not cut). It ships as the `compliance-hub` binary via the flake
  (`nix run .#compliance-hub`). STACK.md records the decision — do not
  re-litigate it in a release run.
- **Tags are cut only when a code change justifies them** — no cosmetic
  releases for release-page optics.

## Pre-tag sequence (in order, every step with a TRUE exit code)

1. **Cut the CHANGELOG section** (`## [X.Y.Z] - YYYY-MM-DD`) and write the
   release notes file (`docs/notes/vX.Y.Z-release-notes.md`, copied from the
   section). The in-repo notes file stays the single source for the GitHub
   release body.
2. **Wait for a clean tree** (the auto-commit daemon lands everything) and a
   ~10-minute quiet window (parallel-session quiescence — the css artifacts
   ping-ponged under a concurrent session during v0.2.0).
3. **Watch CI green on the tip commit** BEFORE tagging:
   `gh run watch <run-id> --exit-status` then read `$?` outside any pipe.
   The v0.2.0 tag froze a 13/14 red run because a pipe-masked watch was
   believed green. Never repeat that.
4. **Run `scripts/pre-release-check.sh vX.Y.Z`** — it re-checks the tree,
   the CHANGELOG, the notes file, master-pushed-on-BOTH-remotes, tag
   freshness, the CANONICAL gate (plain `buildflow`), `nix flake check`,
   and `docs-gates run .`.
   - The canonical gate is plain `buildflow`, not a per-tool sweep: a
     manual-green drill-down missed 8 error findings during the v0.2.0
     prep. Never overlap concurrent full buildflow runs (SQLITE_BUSY).
5. **Tag and push BOTH remotes WITH the tag**:
   `git tag -a vX.Y.Z -m "go-compliance vX.Y.Z"` then
   `git push origin master vX.Y.Z && git push github master vX.Y.Z`.
6. **GitHub release**:
   `gh release create vX.Y.Z --repo LarsArtmann/go-compliance --verify-tag
   --prerelease --notes docs/notes/vX.Y.Z-release-notes.md`
   (v0.x releases carry `--prerelease`; drop it when cutting 1.0+.)

## Post-tag sequence

1. **Verify end-to-end**: `scripts/pre-release-check.sh --verify vX.Y.Z`
   (both remotes at the tagged commit, proxy `@v/list` contains the tag,
   GitHub release page live, clean-dir `go get` resolves).
2. **Drop the directory replace if the hub consumes core** (when core's
   module path version moved): `require vX.Y.Z`, `go mod tidy` in `hub/`,
   re-bootstrap BOTH flake vendorHash sites (the hub hash lives at two
   flake sites), `nix flake check` green.
3. **Record the release** in CHANGELOG `[Unreleased]` only for post-tag
   work; the tagged section is immutable history.

## Known sharp edges

- **vendorHash after ANY root-module Go-file edit** (and after go.mod
  bumps): the failure signature is a hash-mismatch FOD error cascading
  into every `packages.*`/`checks.*`. Bootstrap with `lib.fakeHash`, run
  `nix build`, paste the `got:` hash, re-run (documented ritual; the
  devShell and hub hashes are separate sites).
- **css artifacts move with the tailwind pin**: `scripts/regen-ui.sh` with
  the LOCKED tool (`nix shell .#tailwindcss`) and commit the three
  artifacts in the same commit as any pin/lock change. CI diffs against a
  pinned nixpkgs rev URL — registry-resolved tools drift under identical
  version strings.
- **Tag-triggered CI** re-runs the matrix on `v*` pushes (added
  2026-11-24); the tag's own CI verdict is the release verdict.
- **v0.1.0 ships tag-only** (zero notes, no GitHub release page until the
  2026-11-24 retro-release); v0.2.0's GitHub page is a `--prerelease`.
  The releases page is part of the release, not an afterthought.
