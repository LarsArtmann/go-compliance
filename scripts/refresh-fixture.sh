#!/usr/bin/env bash
# refresh-fixture.sh regenerates the golden v1 wire fixture from a live
# BuildFlow producer. Repository names are redacted (the fixture ships in a
# public repo); everything else is real producer output. Verify the diff
# before committing: the fixture pins the frozen contract.
set -euo pipefail

cd "$(dirname "$0")/.."

buildflow compliance --json | go run ./scripts/redactfixture >testdata/buildflow-v1-fixture.json

echo "OK: testdata/buildflow-v1-fixture.json refreshed (redacted)"
