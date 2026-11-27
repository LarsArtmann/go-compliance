#!/usr/bin/env bash
# screenshot-page.sh — one static viewport capture of a page or a local HTML
# file, taken with the same pinned chromium as scripts/screenshot.sh. For
# documentation assets (docs/assets); the dashboard gate stays
# scripts/screenshot.sh.
#
# Usage: scripts/screenshot-page.sh <url | path-to-html> <out.png> [WIDTHxHEIGHT]
# Environment: CHROMIUM_BIN.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

target="${1:-}"
out="${2:-}"
size="${3:-1440x1000}"

[ -n "$target" ] && [ -n "$out" ] || {
	echo "usage: scripts/screenshot-page.sh <url | path-to-html> <out.png> [WIDTHxHEIGHT]" >&2
	exit 2
}

case "$target" in
http://* | https://* | file://*) url="$target" ;;
*)
	[ -f "$target" ] || {
		echo "✗ no file at '$target'" >&2
		exit 1
	}
	url="file://$root/$target"
	;;
esac

if [ -z "${CHROMIUM_BIN:-}" ]; then
	CHROMIUM_BIN="$(nix build "github:NixOS/nixpkgs/$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)#chromium" \
		--print-out-paths --no-link)/bin/chromium"
fi
[ -x "$CHROMIUM_BIN" ] || {
	echo "✗ no chromium at '$CHROMIUM_BIN' — set CHROMIUM_BIN to a chromium binary" >&2
	exit 1
}

profile="$(mktemp -d)"
trap 'rm -rf "$profile"' EXIT

mkdir -p "$(dirname "$out")"
"$CHROMIUM_BIN" --headless=new --no-sandbox --disable-gpu --hide-scrollbars \
	--user-data-dir="$profile" --window-size="$size" --virtual-time-budget=4000 \
	--screenshot="$root/$out" "$url" >/dev/null 2>&1

[ -s "$out" ] || {
	echo "✗ chromium wrote nothing to $out" >&2
	exit 1
}

echo "✓ captured $url ($size) into $out"
