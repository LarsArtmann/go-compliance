#!/usr/bin/env bash
# regen-ui.sh — regenerate the hub dashboard's tracked CSS artifacts in one
# command (the canonical pipeline; CI regenerates and diffs to keep them
# byte-fresh — an adopted component whose utilities never reached the
# stylesheet renders unstyled in production).
#
# Run after touching any dashboard .templ file, adopting a templ-components
# component, or bumping the templ-components pin, then commit the artifacts
# together with the change that motivated them:
#   hub/internal/dashboard/library-classes.txt
#   hub/internal/dashboard/library-custom.css
#   hub/internal/dashboard/app.min.css
#
# Tools: tailwindcss (v4, pinned by the devShell / CI); go (for the
# module-cache resolution in gen-library-classes.sh).
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
cd "$root"

for tool in tailwindcss go; do
  command -v "$tool" >/dev/null || {
    echo "✗ missing tool '$tool' — run inside 'nix develop' (the devshell pins the version CI verifies)" >&2
    exit 1
  }
done

echo "• library class inventory"
bash ./scripts/gen-library-classes.sh

echo "• tailwind (minified)"
tailwindcss -i hub/internal/dashboard/app.css -o hub/internal/dashboard/app.min.css --minify

echo "✓ regenerated — commit the three artifacts above"
