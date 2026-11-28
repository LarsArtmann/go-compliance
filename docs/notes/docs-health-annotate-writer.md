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

## Formatter rules for swept reports (2026-11-28, reproduced on the real files)

dprint does not read `~~N.` as a list item. A struck item that directly
touches a bare neighbour damages the whole list on the next `dprint fmt`:

- a struck item directly BELOW a bare item is nested into it (indented),
  and every following item is renumbered;
- a struck item directly ABOVE a bare item costs the spaces around inline
  code in that neighbour (`block`$?`reads`);
- an item written as `22.`--flag`` (no space after the dot) is not an
  item at all and becomes a continuation line.

So: a blank line above and below every struck item; a multi-line item
strikes its whole block (`~~head` ... `last line~~ done — verdict`), the
verdict ends the item; never split a verdict over the item's first line.

The mechanical check and repair is `node scripts/annotate-qa.mjs --base
add [--fix] <file>...` in go-compliance (compares with the commit that
added the file). Run it after every `dprint fmt` on a swept report and
before reporting a lane done; exit 0 means no finding remains.
