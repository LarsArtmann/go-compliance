# docs-health annotate writer notes

Relocated from AGENTS.md 2026-11-20 (canonical-gate line-count rule); content unchanged.

- Keys into reports with MULTIPLE numbered sections MUST use the
  `N@substring` disambiguator — a number-only key strikes the FIRST
  matching item number across sections (one repair cycle burned
  2026-11-19). Always `emit-keys` + `--dry-run` the first spec against
  an unfamiliar file shape.
- The writer (`go run
  /home/lars/.config/crush/skills/docs-health/assets/annotate.go`) owns
  the row gates: `repair <tsv>` (spec `path\tlines\tkind\tvalue`; kinds
  h/v/p/w/n/r/s; line-addressed, dry-run first, atomic), `check-rows
  [--strict]`, `bare-rows`, `emit-keys`, `status-items`, `self-test`.
  The installed `docs-gates` binary (fleet pin `e305834`, verified
  2026-11-20) lacks ONLY the `rows` verb — its freshness surface is
  complete (`--pre-flight`, `--margins`, `--dump-items`,
  `--explain-item`, `--audit-all`, `--audit-max`, `--replace`,
  `--duplicates`, `--land-parity`). Use the writer aliases for the row
  gates until the fleet pin bumps; battery stays `docs-gates run .`
  (5 gates).
