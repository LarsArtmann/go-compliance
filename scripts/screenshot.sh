#!/usr/bin/env bash
# screenshot.sh — capture the fleet dashboard in light, dark, and narrow
# (phone-width) views. scripts/screenshot.mjs drives a headless chromium over
# the DevTools protocol and fails the run on console errors, a theme that did
# not apply, an empty fleet table, or horizontal overflow.
#
# Without a URL argument the hub is built from the working tree and serves
# the XDG evidence file on a loopback port, so uncommitted template changes
# can be reviewed before they ship. Pass a URL to capture a running hub.
#
# Usage: scripts/screenshot.sh [url]
# Environment: OUT (default /tmp/compliance-screens), CHROMIUM_BIN,
# SHOTS_EVIDENCE (default ~/.local/state/overview/compliance.json),
# SHOTS_PORT (default 8479), SETTLE_MS (default 2500), ALLOW_EMPTY=1.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

OUT="${OUT:-/tmp/compliance-screens}"
URL="${1:-}"
PORT="${SHOTS_PORT:-8479}"
MIN_PNG_BYTES=20000

if [ -z "${CHROMIUM_BIN:-}" ]; then
	CHROMIUM_BIN="$(nix build "github:NixOS/nixpkgs/$(jq -r '.nodes.nixpkgs.locked.rev' flake.lock)#chromium" \
		--print-out-paths --no-link)/bin/chromium"
fi
export CHROMIUM_BIN
if [ ! -x "$CHROMIUM_BIN" ]; then
	echo "✗ no chromium at '$CHROMIUM_BIN' — set CHROMIUM_BIN to a chromium binary" >&2
	exit 1
fi

HUB_PID=""
cleanup() {
	if [ -n "$HUB_PID" ]; then
		kill "$HUB_PID" 2>/dev/null || true
	fi
	if [ -n "${TMP_DIR:-}" ]; then
		rm -rf "$TMP_DIR"
	fi
}
trap cleanup EXIT

if [ -z "$URL" ]; then
	EVIDENCE="${SHOTS_EVIDENCE:-$HOME/.local/state/overview/compliance.json}"
	if [ ! -f "$EVIDENCE" ]; then
		echo "✗ no evidence file at '$EVIDENCE' — set SHOTS_EVIDENCE or pass a running hub URL" >&2
		exit 1
	fi

	TMP_DIR="$(mktemp -d)"
	(cd hub && GOWORK=off go build -o "$TMP_DIR/compliance-hub" ./cmd/compliance-hub)
	"$TMP_DIR/compliance-hub" serve --source file --file "$EVIDENCE" --addr "127.0.0.1:$PORT" >"$TMP_DIR/hub.log" 2>&1 &
	HUB_PID=$!

	ready=0
	for _ in $(seq 1 100); do
		if (exec 3<>"/dev/tcp/127.0.0.1/$PORT") 2>/dev/null; then
			ready=1
			break
		fi
		sleep 0.2
	done
	if [ "$ready" -ne 1 ]; then
		echo "✗ the hub built from the working tree never listened on 127.0.0.1:$PORT" >&2
		cat "$TMP_DIR/hub.log" >&2
		exit 1
	fi

	URL="http://127.0.0.1:$PORT/"
fi

mkdir -p "$OUT"
node scripts/screenshot.mjs "$URL" "$OUT"

for name in dashboard-light dashboard-dark dashboard-narrow; do
	path="$OUT/$name.png"
	if [ ! -s "$path" ] || [ "$(wc -c <"$path")" -lt "$MIN_PNG_BYTES" ]; then
		echo "✗ $name.png is missing or trivially small — a blank or error page, not the dashboard" >&2
		exit 1
	fi
done

echo "✓ captured $URL into $OUT"
