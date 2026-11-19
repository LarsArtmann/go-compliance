#!/usr/bin/env bash
# check-citations.sh — mechanical verification of file:line citations in the
# library-utilization reports (docs/review/libs/*.md).
#
# What it does: extracts every `file.ext:LINE` / `file.ext:LINE-LINE`
# citation and checks it against the repository tree. A citation whose file
# resolves in this repo must have its cited line(s) inside the file's
# current bounds — the exact drift class the 2026-11 Wave refactors
# (file splits, handler extraction, package moves) produce.
#
# Resolution rules:
#   - `path/with/slashes.go:12` must exist at that repo-relative path
#     (a path citation that vanished is stale by definition).
#   - bare `file.go:12` resolves against every in-repo file with that
#     basename; the citation passes if ANY match still covers the range
#     (prose usually names just the basename).
#   - a bare basename with NO in-repo match is reported as presumed
#     LIBRARY-side cite (the reports quote the audited lib's own sources
#     too); info only, never a failure.
#
# Usage:  scripts/check-citations.sh [report.md ...]   (default: all reports)
# Output: TSV lines  <status>\t<doc>\t<citation>\t<detail>
# Exit:   0 = no stale citations, 1 = stale found, 2 = usage error.
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

if [ $# -gt 0 ]; then
  targets=("$@")
else
  targets=(docs/review/libs/*.md)
fi

CITE_RE='[A-Za-z0-9_][A-Za-z0-9_./-]*\.(go|templ|md|yml|yaml|json|nix|sh|sql|css|ts|js):[0-9]+(-[0-9]+)?'

declare -A BY_BASENAME
while IFS= read -r f; do
  b="${f##*/}"
  BY_BASENAME[$b]="${BY_BASENAME[$b]:-}$f"$'\n'
done < <(git ls-files)

stale=0
checked=0
external=0
ok=0

check_one() { # $1=doc $2=citation
  local doc="$1" cite="$2" file part start end
  file="${cite%%:*}"
  part="${cite#*:}"
  start="${part%%-*}"
  if [[ "$part" == *-* ]]; then end="${part#*-}"; else end="$start"; fi

  local -a candidates=()
  if [[ "$file" == */* ]]; then
    [ -f "$file" ] && candidates=("$file")
  else
    while IFS= read -r m; do
      [ -n "$m" ] && candidates+=("$m")
    done <<<"${BY_BASENAME[$file]:-}"
  fi

  if [ ${#candidates[@]} -eq 0 ]; then
    printf 'EXTERNAL\t%s\t%s\tno in-repo file; presumed library-side cite\n' "$doc" "$cite"
    external=$((external + 1))
    return
  fi

  local best="" c lines
  for c in "${candidates[@]}"; do
    lines=$(wc -l <"$c")
    if [ "$start" -le "$lines" ] && [ "$end" -le "$lines" ]; then
      best="$c ($lines lines)"
      break
    fi
  done
  if [ -n "$best" ]; then
    ok=$((ok + 1))
  else
    printf 'STALE\t%s\t%s\tresolved candidates out of range: %s\n' \
      "$doc" "$cite" "$(printf '%s ' "${candidates[@]}")"
    stale=$((stale + 1))
  fi
  checked=$((checked + 1))
}

for doc in "${targets[@]}"; do
  [ -f "$doc" ] || { echo "no such report: $doc" >&2; exit 2; }
  while IFS= read -r cite; do
    [ -n "$cite" ] && check_one "$doc" "$cite"
  done < <(rg -o "$CITE_RE" "$doc" | sort -u)
done

echo "--"
echo "checked=$checked ok=$ok stale=$stale external(presumed lib-side)=$external"
[ "$stale" -eq 0 ]
