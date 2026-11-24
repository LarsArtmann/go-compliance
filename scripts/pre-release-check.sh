#!/usr/bin/env bash
# Release guard for go-compliance. Encodes the lessons from the v0.2.0
# release run (2026-11-20):
#
#   1. Pipeline exit codes LIE when piped (`cmd | tail` masks failures —
#      3 strikes in one session). This script never pipes a gated command.
#   2. The CANONICAL gate is plain `buildflow` — a manual per-tool sweep
#      that is green is not release-ready (the same session nearly tagged
#      with 8 buildflow error findings hidden behind a green drill-down).
#   3. Push master to BOTH remotes (Forgejo origin + GitHub mirror) BEFORE
#      tagging, and never tag a commit whose CI has not run green.
#   4. The release needs a CHANGELOG section and a release-notes file.
#   5. Verify the released artifact afterwards: the module proxy must
#      serve the tag and the GitHub release must exist.
#
# Usage:
#   scripts/pre-release-check.sh <version>          # pre-flight (before tagging)
#   scripts/pre-release-check.sh --verify <version> # post-release verification
#
# <version> is the bare tag name, e.g. v0.2.1.
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

step() { printf "${GREEN}==>%s %s${NC}\n" "" "$1"; }
warn() { printf "${YELLOW}WARNING: %s${NC}\n" "$1"; }
die() { printf "${RED}FAIL: %s${NC}\n" "$1" >&2; exit 1; }

usage() {
	echo "usage: $0 <version> | $0 --verify <version>"
	echo "  example: $0 v0.2.1"
	exit 2
}

MODE="preflight"
VERSION=""
case "${1:-}" in
	--verify)
		MODE="verify"
		VERSION="${2:-}"
		;;
	v*)
		VERSION="${1:-}"
		;;
	*)
		usage
		;;
esac
[ -n "$VERSION" ] || usage

if [ "$MODE" = "verify" ]; then
	step "Checking the tag exists on BOTH remotes at the same commit"
	local_rev="$(git rev-parse "$VERSION")"
	for remote in origin github; do
		remote_rev="$(git ls-remote --tags "$remote" "refs/tags/$VERSION" | cut -f1)"
		if [ -z "$remote_rev" ]; then
			die "$VERSION is missing on $remote — push the tag there first"
		fi
		if [ "$remote_rev" != "$local_rev" ]; then
			die "$VERSION on $remote ($remote_rev) != local ($local_rev)"
		fi
		echo "    $remote == $local_rev"
	done

	step "Checking the module proxy serves the tag"
	proxy_info="$(go list -m -versions github.com/larsartmann/go-compliance 2>/dev/null)"
	case "$proxy_info" in
		*"$VERSION"*) echo "    proxy @v/list contains $VERSION" ;;
		*) die "proxy @v/list does not contain $VERSION (got: $proxy_info)" ;;
	esac

	step "Checking the GitHub release exists"
	if gh release view "$VERSION" --repo LarsArtmann/go-compliance >/dev/null 2>&1; then
		echo "    release page live"
	else
		warn "no GitHub release page for $VERSION — create it (gh release create $VERSION --verify-tag --notes docs/notes/${VERSION}-release-notes.md)"
	fi

	step "Checking a clean-dir consumer resolve"
	tmpdir="$(mktemp -d)"
	if (cd "$tmpdir" && GOWORK=off GOFLAGS=-mod=mod go mod init consumer-verify >/dev/null 2>&1 && GOWORK=off go get "github.com/larsartmann/go-compliance@$VERSION" >/dev/null 2>&1); then
		echo "    go get @$VERSION exit 0"
	else
		die "clean-dir go get @$VERSION failed"
	fi
	trash "$tmpdir" 2>/dev/null || true

	echo "${GREEN}VERIFY OK: $VERSION is live end-to-end.${NC}"
	exit 0
fi

notes_file="docs/notes/${VERSION}-release-notes.md"

step "Checking working tree is clean"
if ! git diff --quiet || ! git diff --cached --quiet; then
	die "working tree has uncommitted changes - commit (or wait for the auto-commit daemon) first"
fi
echo "    clean"

step "Checking CHANGELOG.md has a [$VERSION] section"
if ! grep -q "^## \[$VERSION\]" CHANGELOG.md; then
	die "CHANGELOG.md has no '## [$VERSION]' section - cut one first"
fi
echo "    section present"

step "Checking release notes file"
if [ ! -f "$notes_file" ]; then
	warn "$notes_file missing - create it (copy from the CHANGELOG [$VERSION] section) before tagging"
else
	echo "    $notes_file"
fi

step "Checking master is pushed to BOTH remotes (tag AFTER push)"
local_head="$(git rev-parse HEAD)"
for remote in origin github; do
	git fetch "$remote" master >/dev/null 2>&1 || die "cannot fetch master from $remote"
	remote_head="$(git rev-parse "$remote/master")"
	if [ "$remote_head" != "$local_head" ]; then
		die "local master ($local_head) != $remote/master ($remote_head) - push first, tag second"
	fi
	echo "    $remote/master == $local_head"
done

step "Checking the tag does not exist yet"
if git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null; then
	die "tag $VERSION already exists locally"
fi
for remote in origin github; do
	if [ -n "$(git ls-remote --tags "$remote" "refs/tags/$VERSION")" ]; then
		die "tag $VERSION already exists on $remote"
	fi
done
echo "    $VERSION is fresh"

step "Running the CANONICAL gate: plain buildflow (real exit code, no pipe)"
if ! buildflow; then
	die "buildflow failed - fix ALL findings before releasing (manual per-tool greens are NOT a substitute)"
fi
echo "    buildflow green"

step "Running nix flake check (real exit code, no pipe)"
if ! nix flake check; then
	die "nix flake check failed - fix ALL checks before releasing (vendorHash goes stale after go.mod changes)"
fi
echo "    all checks green"

step "Running docs-gates run . (real exit code, no pipe)"
if ! docs-gates run .; then
	die "docs-gates failed - write the release status report / harvest open items first"
fi
echo "    docs gates green"

cat <<RUNBOOK

${GREEN}PRE-FLIGHT OK — release runbook for $VERSION:${NC}

  1. Commit any last changes; re-run this script (it wants a clean tree).
  2. Watch CI green on the tip commit BEFORE tagging (gh run watch with
     --exit-status, then read \$? outside any pipe).
  3. Tag: git tag -a $VERSION -m "go-compliance $VERSION"
  4. Push BOTH remotes WITH the tag:
       git push origin master "$VERSION" && git push github master "$VERSION"
  5. GitHub release (notes file):
       gh release create $VERSION --repo LarsArtmann/go-compliance \\
         --verify-tag --prerelease \\
         --notes "$notes_file"
  6. Post-release verification: scripts/pre-release-check.sh --verify $VERSION
RUNBOOK
