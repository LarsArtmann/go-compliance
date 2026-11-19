# Security Policy

## Supported versions

| Version | Supported          |
| ------- | ------------------ |
| dev     | :white_check_mark: |

Pre-1.0: the repository is under active development before the first tagged
release; the tip of `master` is the only supported state. Once the core
module tags v0.1.0, only the latest minor receives security fixes.

## Reporting a vulnerability

Please do **not** open a public issue for security reports.

- Use GitHub private vulnerability reporting (Security → Report a
  vulnerability), which keeps the report confidential until a fix ships.
- Alternatively contact the maintainer directly on GitHub (@LarsArtmann).

Include what you can of: the affected version
(`go list -m github.com/larsartmann/go-compliance`), a minimal reproduction,
and your assessment of impact. You will get an acknowledgement within a
week, and a fix or a mitigation plan before any public disclosure.

## Security Practices

- **Bind posture is localhost-only by default** (`127.0.0.1:8477`; unix
  sockets are same-machine by construction). Binding a non-loopback address
  prints an explicit warning and exposes fleet evidence to the network —
  that is a deliberate operator decision, not a bug. See
  [docs/decisions/bind-posture.md](docs/decisions/bind-posture.md).
- **No authentication surface exists yet.** The dashboard, SSE stream, and
  JSON API trust the network they are bound to. Until an auth layer lands,
  the only safe exposure is loopback or a unix socket with restricted
  permissions.
- **Evidence is operator-local data** (repo names, CI verdicts, finding
  counts). The redaction posture of the frozen v1 fixture lives in
  `scripts/refresh-fixture.sh`; producers own what they emit.
- **The SSE surfaces are framing-safe**: the discovery watch uses the
  WHATWG-conformant `daemon.ParseSSEData` with a 1 MiB line cap (over-cap
  frames fail loudly, never truncate), and the dashboard stream sends
  DataStar patches only — payload decoding stays in typed Go.
- `X-Content-Type-Options: nosniff` is set on every response; the embedded
  DataStar bundle is served with an ETag and day-long caching under a
  version-pinned URL.

## Verification

- `govulncheck ./...` must report zero reachable vulnerabilities in both
  modules; the `go` toolchain directive is bumped past any stdlib advisory
  that lands (currently `go1.27.4`, closing GO-2026-6859).
- `nix flake check` builds and tests both modules in the sandbox.
