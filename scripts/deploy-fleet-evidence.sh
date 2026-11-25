#!/usr/bin/env bash
# deploy-fleet-evidence.sh — build library-policy from its COMMITTED tree,
# GC-root the store path, install the daily evidence timer, and, with --run,
# record one fleet scan now and prove the evidence file changed.
#
# Usage: scripts/deploy-fleet-evidence.sh [--run]
# Environment: LIBRARY_POLICY_REPO (default ~/projects/library-policy),
# OUT_LINK (default ~/.local/state/compliance-fleet/current),
# EVIDENCE_FILE (default ~/.local/state/overview/compliance.json).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

repo="${LIBRARY_POLICY_REPO:-$HOME/projects/library-policy}"
out_link="${OUT_LINK:-$HOME/.local/state/compliance-fleet/current}"
evidence="${EVIDENCE_FILE:-$HOME/.local/state/overview/compliance.json}"
unit_dir="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"

fail() {
	echo "✗ $*" >&2
	exit 1
}

[ "$(id -u)" -ne 0 ] || fail "run as your own user, never root: root cannot fetch the private Go dependencies"
[ -d "$repo/.git" ] || fail "no library-policy checkout at '$repo' (set LIBRARY_POLICY_REPO)"

rev="$(git -C "$repo" rev-parse HEAD)"
mkdir -p "$(dirname "$out_link")"
nix build "git+file://$repo?rev=$rev#library-policy" --out-link "$out_link"

bin="$out_link/bin/library-policy"
[ -x "$bin" ] || fail "no library-policy binary under $out_link"
version="$("$bin" --version 2>&1 | tail -1)"

for unit in fleet-evidence.service fleet-evidence.timer; do
	install -Dm 0644 "deploy/$unit" "$unit_dir/$unit"
	verify_output="$(systemd-analyze --user verify "$unit_dir/$unit" 2>&1 || true)"
	if [ -n "$verify_output" ]; then
		fail "systemd-analyze rejected $unit:
$verify_output"
	fi
done

systemctl --user daemon-reload
systemctl --user enable fleet-evidence.timer >/dev/null 2>&1

echo "✓ library-policy $version"
echo "  binary:  $(readlink -f "$bin")"
echo "  units:   $unit_dir/fleet-evidence.service, $unit_dir/fleet-evidence.timer"
echo "  journal: journalctl --user -u fleet-evidence -n 30"

if [ "${1:-}" = "--run" ]; then
	before="$(stat -c %Y "$evidence" 2>/dev/null || echo 0)"
	systemctl --user start fleet-evidence.service
	after="$(stat -c %Y "$evidence" 2>/dev/null || echo 0)"

	if [ "$after" -le "$before" ]; then
		journalctl --user -u fleet-evidence.service -n 30 --no-pager >&2 || true
		fail "the scan ran but the evidence file did not change: $evidence"
	fi

	echo "  run:     evidence recorded to $evidence ($(date -d "@$after" '+%H:%M:%S'))"
fi
