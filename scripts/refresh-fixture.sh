#!/usr/bin/env bash
# refresh-fixture.sh regenerates the golden v1 wire fixture from a live
# BuildFlow producer. Repository names are redacted (the fixture ships in a
# public repo); everything else is real producer output. Verify the diff
# before committing: the fixture pins the frozen contract.
#
# REFRESH RITUAL — the fixture currently pins 341 repos, and three things
# are keyed to that count:
#   1. v1_golden_test.go — byte-exact decode/marshal round-trip; ANY byte
#      change (new repos, redaction drift, key reorder) must be eyeballed
#      against the producer diff before committing.
#   2. bench_test.go — the @341 benchmark shape; a changed count means the
#      0341repos rows measure a different fleet.
#   3. benchmarks/baseline.txt — recorded against the current fixture;
#      after a refresh AND a count change, re-record:
#        go test -run '^$' -bench . -benchmem -count=1 . > benchmarks/baseline.txt
#      (quiet machine; the baseline is a comparison anchor, not a gate).
# Freshness guard (XP2 follow-up, 2026-11-19): refresh ONLY from a binary
# you just built - verify `buildflow --version` matches a fresh `nix build .`
# in the BuildFlow checkout first. A stale PATH binary silently captures
# OLD contract bytes and the golden test then pins stale output.
# After any refresh: run the full root-module test suite; the golden test
# failing means the contract moved — decide deliberately, not by reflex.
set -euo pipefail

cd "$(dirname "$0")/.."

buildflow compliance --json | go run ./scripts/redactfixture >testdata/buildflow-v1-fixture.json

echo "OK: testdata/buildflow-v1-fixture.json refreshed (redacted)"
