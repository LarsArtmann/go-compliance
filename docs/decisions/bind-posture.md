# Decision: bind posture

Date: 2026-11-17
Status: decided (default); open question g1 (auth) may reopen it.
Appendix A (the user-unit deployment) and the auth sketch were added on
2026-11-25; the LAN deployment is live and waits for the g1 sign-off.

## Context

The hub serves fleet compliance evidence: repository names, gate outcomes,
finding counts. That is operational metadata about every LarsArtmann
project — not secrets, but not public either. The overview dashboard it
replaces ran localhost-only behind an SSH tunnel, and its federation
boundary handled remote cases.

Open owner question **g1**: should the hub ever grow authentication, or is
it permanently a localhost/tunnel service? Until answered, the safe default
stands.

## Decision

1. **Default bind is `127.0.0.1:8477`.** No flag, no exposure.
2. **Non-localhost binds are explicit and loud**: passing `--addr
   0.0.0.0:8477` works but prints a warning naming this document. The
   flag is the audit trail; the warning is the speed bump.
3. **Unix-socket listen mode is first-class**: `--addr unix:/run/
   compliance/hub.sock` binds a filesystem socket (dir created if
   missing, socket removed on shutdown). This is the intended
   same-machine integration path — the project-discovery daemon already
   speaks this pattern, and socket permissions replace TCP auth entirely
   (owner-only by default).
4. **No authentication layer until g1 is answered.** Adding auth without
   a threat model is ceremony; the localhost/socket defaults are the
   boundary.

## Consequences

- Same-machine consumers (overview, federation peers via SSH forward)
  use the socket or the loopback TCP port interchangeably.
- Systemd unit ships the TCP loopback default; socket mode is one flag
  away and pairs with `RuntimeDirectory=compliance`.
- If g1 lands on "auth", the decision to extend is additive: a middleware
  in front of the mux, default still loopback.

## Appendix A: the deployment on this host (2026-11-25)

The hub runs as a systemd USER unit, installed by `scripts/deploy-hub.sh`:

1. It builds the COMMITTED tree
   (`nix build "git+file://<repo>?rev=<HEAD>#compliance-hub"`), so the
   binary matches a commit exactly and uncommitted work never ships.
2. The result is GC-rooted at `~/.local/state/compliance-hub/current`.
3. `deploy/compliance-hub.user.service` is rendered with the concrete store
   path into `~/.config/systemd/user/compliance-hub.service`, checked with
   `systemd-analyze verify`, enabled, and restarted.
4. It waits for `GET /v1/compliance` to answer 200 on the LAN address and
   prints the binary, the unit, the page, and the journal command.

A redeploy is the same script again. The hub reports its commit as the
version (`compliance-hub --version`, `/version`, and the `buildflow_version`
of the document it serves).

Evidence-file locations have their single home in the [integration guide's locations table](../guides/integrate.md#where-the-evidence-lives)
(system `/var/lib/overview/compliance.json`, user
`~/.local/state/overview/compliance.json`, `GO_COMPLIANCE_EVIDENCE_FILE`
override, per-command `--file` on the hub and `--evidence-file` on
`library-policy`); on this host the user unit reads the XDG path, its
writer is the `fleet-evidence` timer (`library-policy fleet`), and a
system deployment's writer is the BuildFlow daemon piping
`compliance --json`.

The unit runs `serve --source file --file
~/.local/state/overview/compliance.json --addr 192.168.1.150:8477` as the
login user on a read-only filesystem (`ProtectSystem=strict`,
`ProtectHome=read-only`), notifies readiness over sd_notify, feeds a 30s
watchdog, and restarts without a start limit. The evidence file is written
by the `fleet-evidence` timer (`scripts/deploy-fleet-evidence.sh`), and the
hub refreshes 1-2 seconds after each write.

Reachability: the bind is the host's wired LAN address, so only clients
that route to `192.168.1.150` reach it; Wi-Fi and VPN addresses of this host
stay unbound. No firewall change is needed: SystemNix sets
`trustedInterfaces = ["eno1"]`, so LAN traffic bypasses nftables. dnsblockd
resolves `evo-x2.home.lan`, and the page is
`http://evo-x2.home.lan:8477/` from the LAN.

Verify after every change:

```bash
systemctl --user status compliance-hub.service
journalctl --user -u compliance-hub -n 20
scripts/screenshot.sh http://192.168.1.150:8477/
#   four views: layout each view must produce, table scroll and column-header
#   pin, keyboard walk with the toggle's aria-checked trail, filter clicks
#   judged against the rendered fleet, WCAG AA contrast, the empty-fleet run
#   (ALLOW_EMPTY=1), console errors minus classified browser noise.
```

The units are plain user files, not home-manager managed; moving them (and
the library-policy pin) into SystemNix is an owner decision.

## Auth posture: threat sketch and recommendation (awaiting sign-off)

This section answers open question g1 for the deployment above. An agent
cannot accept a risk: the acceptance is the owner's.

**Who can reach it.** Every device that can route to `192.168.1.150:8477`:
wired and Wi-Fi clients of the home LAN, VMs and containers on the host,
and anything a LAN device forwards. Not reachable from outside the LAN
unless a router forwards the port.

**What it exposes.** Read-only fleet metadata: repository names, compliance
states, findings and failure counts, last-run times, the hub version, and
per-repo `gate_findings` (which flagged libraries a repo depends on). The
evidence document carries no file paths, no source, no credentials
(checked against the live file on 2026-11-25).

**What it accepts.** Only `GET` is served: the page, the SSE patch stream,
the SDK bundle, the stylesheet, the icon, the health routes, and
`/v1/compliance`; every other method and every unknown path gets the styled
404. The one query parameter, `?state=`, is checked against the known
states;
templ escapes every interpolated value; no request data is stored or
reflected. There is no authentication.

**What it writes.** Nothing. The unit's filesystem is read-only and the file
source only reads.

**What can go wrong.** Disclosure of the metadata above to a LAN peer; a
local client can hold many SSE connections or flood requests (denial of
service for other viewers). `MemoryMax=256M` bounds the first, and
`Restart=on-failure` restores the process after a failure caused by the
second. Integrity cannot be attacked through this surface.

**Recommendation.** Accept the LAN-trust posture for the home LAN and keep
the hub read-only. Revisit, with unix-socket mode plus an authenticating
reverse proxy or an allowlist middleware, when any of these happens: the
hub becomes reachable from outside the home LAN; it gains a write path (the
v2 reporter ingress); or the evidence grows more sensitive than library
names and counts.

**Sign-off.** Open. `TODO_LIST.md` carries the owner ask.
