# Decision: bind posture

Date: 2026-11-17
Status: decided (default); open question g1 (auth) may reopen it

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
