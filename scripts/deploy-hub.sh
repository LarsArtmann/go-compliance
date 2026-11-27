#!/usr/bin/env bash
# deploy-hub.sh — build compliance-hub from the COMMITTED tree, GC-root the
# store path, render the systemd user unit onto it, and (re)start the
# service. The unit names the concrete store path, so the running binary
# never drifts with the checkout; a redeploy is this script again.
#
# Refuses to take the address from a process the unit does not own (the hub
# itself never steals a live sibling's socket).
#
# Usage: scripts/deploy-hub.sh
# Environment: HUB_ADDR (default 192.168.1.150:8477),
# OUT_LINK (default ~/.local/state/compliance-hub/current).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

addr="${HUB_ADDR:-192.168.1.150:8477}"
out_link="${OUT_LINK:-$HOME/.local/state/compliance-hub/current}"
unit_name="compliance-hub.service"
unit_dir="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
template="deploy/compliance-hub.user.service"
port="${addr##*:}"

fail() {
	echo "✗ $*" >&2
	exit 1
}

[ "$(id -u)" -ne 0 ] || fail "run as your own user, never root: the evidence file is user-owned"
command -v nix >/dev/null || fail "nix is not on PATH"
command -v curl >/dev/null || fail "curl is not on PATH (the post-deploy probe needs it)"
command -v jq >/dev/null || fail "jq is not on PATH (the readiness probe needs it)"
[ -f "$template" ] || fail "unit template $template is missing"

rev="$(git rev-parse HEAD)"
mkdir -p "$(dirname "$out_link")"
nix build "git+file://$root?rev=$rev#compliance-hub" --out-link "$out_link"

bin="$(readlink -f "$out_link/bin/compliance-hub")"
version="$("$bin" --version 2>&1 | tail -1)"
reported="${version##* }"
case "$rev" in
"$reported"*) ;;
*)
	fail "the built hub reports '$version' but HEAD is ${rev:0:12}; the flake's version stamp is out of step"
	;;
esac

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
	fail "readiness is not green after 10s (checks not passing: $not_passing)"
fi

unit_state="$(systemctl --user is-active "$unit_name" 2>/dev/null || true)"
listener="$(ss -ltnpH "sport = :$port" 2>/dev/null || true)"
if [ "$unit_state" != "active" ] && [ -n "$listener" ]; then
	fail "another process owns $addr and the unit does not:
$listener
Stop it first; this script does not take the address from a process it did not start."
fi

mkdir -p "$unit_dir"
sed -e "s|@BIN@|$bin|g" -e "s|@ADDR@|$addr|g" "$template" >"$unit_dir/$unit_name"
verify_output="$(systemd-analyze --user verify "$unit_dir/$unit_name" 2>&1 || true)"
if [ -n "$verify_output" ]; then
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
	fail "the hub did not answer 200 on $probe within 20s"
fi

unit_state="$(systemctl --user is-active "$unit_name" 2>/dev/null || true)"
if [ "$unit_state" != "active" ]; then
	systemctl --user status "$unit_name" --no-pager -n 20 >&2 || true
	fail "the unit is '$unit_state' after the restart"
fi

echo "✓ $version"
echo "  binary:  $bin"
echo "  unit:    $unit_dir/$unit_name"
echo "  page:    http://$addr/"
echo "  api:     $probe"
echo "  journal: journalctl --user -u compliance-hub -n 20"
