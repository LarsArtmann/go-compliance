#!/usr/bin/env bash
# deploy-hub.sh — build compliance-hub from the COMMITTED tree, GC-root the
# store path, render the systemd user unit onto it, and (re)start the
# service. The unit names the concrete store path, so the running binary
# never drifts with the checkout; a redeploy is this script again.
#
# The new hub passes three gates before the deploy counts: it answers, every
# readiness check passes, and the real-browser harness (scripts/screenshot.sh)
# is green against the live address. A gate that fails after the restart
# rolls the unit file and the out-link back to the previous hub and starts
# that one again.
#
# Refuses to take the address from a process the unit does not own (the hub
# itself never steals a live sibling's socket).
#
# Usage: scripts/deploy-hub.sh
# Environment: HUB_ADDR (default 192.168.1.150:8477),
# OUT_LINK (default ~/.local/state/compliance-hub/current), and the
# browser harness reads OUT, CHROMIUM_BIN, and SETTLE_MS (see
# scripts/screenshot.sh).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

addr="${HUB_ADDR:-192.168.1.150:8477}"
out_link="${OUT_LINK:-$HOME/.local/state/compliance-hub/current}"
unit_name="compliance-hub.service"
unit_dir="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
unit_file="$unit_dir/$unit_name"
template="deploy/compliance-hub.user.service"
harness="scripts/screenshot.sh"
port="${addr##*:}"

fail() {
	echo "✗ $*" >&2
	exit 1
}

[ "$(id -u)" -ne 0 ] || fail "run as your own user, never root: the evidence file is user-owned"
command -v nix >/dev/null || fail "nix is not on PATH"
command -v curl >/dev/null || fail "curl is not on PATH (the post-deploy probe needs it)"
command -v jq >/dev/null || fail "jq is not on PATH (the readiness probe needs it)"
command -v node >/dev/null || fail "node is not on PATH (the browser harness drives chromium with it)"
[ -f "$template" ] || fail "unit template $template is missing"
[ -f "$harness" ] || fail "browser harness $harness is missing"

rev="$(git rev-parse HEAD)"
mkdir -p "$(dirname "$out_link")"

# Rollback material, read before anything changes: the out-link target (the
# GC root of a hub that may be running now) and the unit file as rendered.
previous_target="$(readlink "$out_link" 2>/dev/null || true)"
previous_unit=""
if [ -f "$unit_file" ]; then
	previous_unit="$(cat "$unit_file")"
fi

# restore_link puts the GC root back onto the hub that was current before
# this run, so an aborted deploy cannot leave a running binary collectable.
restore_link() {
	if [ -n "$previous_target" ] && [ -e "$previous_target" ]; then
		ln -sfn "$previous_target" "$out_link"
	fi
}

# restore_unit puts the unit file back; the file of a first deploy, which
# has no predecessor, is moved aside for inspection.
restore_unit() {
	if [ -n "$previous_unit" ]; then
		printf '%s\n' "$previous_unit" >"$unit_file"
	elif [ -f "$unit_file" ]; then
		mv "$unit_file" "$unit_file.failed"
	fi
}

# abort fails a deploy whose unit file was not replaced (systemd was not
# told about anything).
abort() {
	restore_link
	fail "$*"
}

# rollback fails a deploy whose unit already took the new hub: the unit file
# and the out-link return to what they were, and a unit that ran before the
# deploy runs again.
rollback() {
	echo "✗ $*" >&2
	echo "  rolling back to the previous hub" >&2
	if [ -z "$previous_unit" ]; then
		systemctl --user disable "$unit_name" >/dev/null 2>&1 || true
	fi
	restore_unit
	restore_link
	systemctl --user daemon-reload
	if [ "$was_active" = "active" ]; then
		systemctl --user restart "$unit_name"
	fi
	exit 1
}

nix build "git+file://$root?rev=$rev#compliance-hub" --out-link "$out_link"

bin="$(readlink -f "$out_link/bin/compliance-hub")"
version="$("$bin" --version 2>&1 | tail -1)"
reported="${version##* }"
case "$rev" in
"$reported"*) ;;
*)
	abort "the built hub reports '$version' but HEAD is ${rev:0:12}; the flake's version stamp is out of step"
	;;
esac

was_active="$(systemctl --user is-active "$unit_name" 2>/dev/null || true)"
listener="$(ss -ltnpH "sport = :$port" 2>/dev/null || true)"
if [ "$was_active" != "active" ] && [ -n "$listener" ]; then
	abort "another process owns $addr and the unit does not:
$listener
Stop it first; this script does not take the address from a process it did not start."
fi

mkdir -p "$unit_dir"
sed -e "s|@BIN@|$bin|g" -e "s|@ADDR@|$addr|g" "$template" >"$unit_file"
verify_output="$(systemd-analyze --user verify "$unit_file" 2>&1 || true)"
if [ -n "$verify_output" ]; then
	restore_unit
	restore_link
	fail "systemd-analyze rejected the rendered unit:
$verify_output"
fi
systemctl --user daemon-reload
systemctl --user enable "$unit_name" >/dev/null 2>&1
systemctl --user restart "$unit_name"

probe="http://$addr/v1/compliance"
ready=0
for _ in $(seq 1 100); do
	if curl -fsS -o /dev/null "$probe" 2>/dev/null; then
		ready=1
		break
	fi
	sleep 0.2
done
if [ "$ready" -ne 1 ]; then
	systemctl --user status "$unit_name" --no-pager -n 20 >&2 || true
	rollback "the new hub did not answer 200 on $probe within 20s"
fi

# Readiness must be fully green, not merely serving: a check stuck on warn
# (a dead discovery stream, a stalled refresh) hides behind a 200.
readyz="http://$addr/readyz"
not_passing=""
for _ in $(seq 1 50); do
	not_passing="$(curl -fsS "$readyz" 2>/dev/null |
		jq -r '[.checks | to_entries[] | select(.value.status != "pass" and .value.status != "off") | .key] | join(", ")' 2>/dev/null || echo "unparsed")"
	[ -z "$not_passing" ] && break
	sleep 0.2
done
if [ -n "$not_passing" ]; then
	curl -fsS "$readyz" 2>/dev/null | jq -r '.checks | to_entries[] | "\(.key): \(.value.status) \(.value.error)"' >&2 || true
	rollback "readiness is not green after 10s (checks not passing: $not_passing)"
fi

unit_state="$(systemctl --user is-active "$unit_name" 2>/dev/null || true)"
if [ "$unit_state" != "active" ]; then
	systemctl --user status "$unit_name" --no-pager -n 20 >&2 || true
	rollback "the unit is '$unit_state' after the restart"
fi

# The real browser is the gate for everything visual: a page that renders
# inert, unreadable, or unusable by keyboard ships to the LAN behind green
# readiness checks, and no probe can see it.
if ! harness_output="$(bash "$harness" "http://$addr/" 2>&1)"; then
	echo "$harness_output" >&2
	rollback "the browser harness is red against the new hub (its findings are above)"
fi

echo "✓ $version"
echo "  binary:  $bin"
echo "  unit:    $unit_file"
echo "  page:    http://$addr/"
echo "  api:     $probe"
echo "  browser: $(printf '%s\n' "$harness_output" | grep '^✓' | tail -1)"
echo "  journal: journalctl --user -u compliance-hub -n 20"
