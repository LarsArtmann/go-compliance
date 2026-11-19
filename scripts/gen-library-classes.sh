#!/usr/bin/env bash
# gen-library-classes.sh — regenerate the templ-components class inventory
# for Tailwind (hub dashboard).
#
# Why: Tailwind v4 refuses to scan .gitignore'd and read-only module-cache
# paths directly, so the library's classes reach the compiler through ONE
# tracked, deterministic file that app.css scans via @source. The inventory
# is rebuilt from go.mod truth every run (go list -m resolves each module
# directory, honoring replace directives), so the scan can never run against
# a stale tree.
#
# Outputs (both tracked, committed alongside the .templ change that
# motivated them):
#   hub/internal/dashboard/library-classes.txt  — concatenated library sources
#   hub/internal/dashboard/library-custom.css   — the lib's templates/custom.css
#                                                  (.tc-* plain-CSS utilities)
#
# Atomicity: the inventory is written to a temp file and moved into place
# only after every module resolved; a mid-run failure leaves the previous
# artifact byte-identical.
#
# Usage: scripts/gen-library-classes.sh   (from the repo root or hub/; needs
# go in PATH and a warm module cache or network for first resolution)
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

dest="$root/hub/internal/dashboard/library-classes.txt"
custom_dest="$root/hub/internal/dashboard/library-custom.css"

# Every templ-components module the hub can render classes from. The root
# module carries layout/display/forms; the submodules carry datastar/,
# errorpage/, icons/, utils/. Sorted for deterministic concatenation order.
modules=(
  github.com/larsartmann/templ-components
  github.com/larsartmann/templ-components/datastar
  github.com/larsartmann/templ-components/errorpage
  github.com/larsartmann/templ-components/icons
  github.com/larsartmann/templ-components/utils
)

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

for m in "${modules[@]}"; do
  dir="$(cd hub && GOWORK=off go list -m -f '{{.Dir}}' "$m")"
  if [ -z "$dir" ] || [ ! -d "$dir" ]; then
    echo "gen-library-classes: cannot resolve $m (go list said: '$dir')" >&2
    exit 1
  fi
  # LC_ALL=C pins the sort so the concatenation is byte-identical across
  # environments. Generated *_templ.go files mirror the .templ sources.
  LC_ALL=C find "$dir" \
    -type f \( -name '*.templ' -o -name '*.go' \) ! -name '*_templ.go' \
    -print0 | LC_ALL=C sort -z | xargs -0 -r cat >>"$tmp"
done

bytes="$(wc -c <"$tmp")"
if [ "$bytes" -eq 0 ]; then
  echo "gen-library-classes: scan produced 0 bytes" >&2
  exit 1
fi
mv "$tmp" "$dest"
trap - EXIT
echo "wrote $dest ($(wc -c <"$dest") bytes)"

# Refresh the library's .tc-* utilities: plain CSS, never emittable by
# class scanning, so the file must be @imported. templates/ is not Go code,
# so the module cache copy is the only source; install (not cp) because the
# cache is read-only and a plain cp would preserve 0444 on the tracked file.
root_dir="$(cd hub && GOWORK=off go list -m -f '{{.Dir}}' github.com/larsartmann/templ-components)"
lib_css="$root_dir/templates/custom.css"
if [ -f "$lib_css" ]; then
  install -m 0644 "$lib_css" "$custom_dest"
  echo "wrote $custom_dest ($(wc -c <"$custom_dest") bytes)"
elif [ ! -f "$custom_dest" ]; then
  echo "gen-library-classes: $lib_css missing and no existing $custom_dest" >&2
  exit 1
else
  echo "kept $custom_dest (module templates/custom.css not present)"
fi
