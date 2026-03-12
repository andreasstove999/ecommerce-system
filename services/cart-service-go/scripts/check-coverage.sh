#!/usr/bin/env bash
set -euo pipefail

THRESHOLD="${COVERAGE_THRESHOLD:-100}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"

packages="$(go list ./... | rg -v '/integration$' || true)"
if [[ -z "${packages}" ]]; then
  echo "No packages found to test"
  exit 1
fi

go test ${packages} -covermode=atomic -coverpkg=./... -coverprofile="${PROFILE}"

total="$(go tool cover -func="${PROFILE}" | awk '/^total:/ {gsub("%", "", $3); print $3}')"
if [[ -z "${total}" ]]; then
  echo "Unable to read total coverage from ${PROFILE}"
  exit 1
fi

awk -v total="${total}" -v threshold="${THRESHOLD}" 'BEGIN { if (total + 0 < threshold + 0) exit 1 }' || {
  echo "Coverage gate failed: ${total}% < ${THRESHOLD}%"
  exit 1
}

echo "Coverage gate passed: ${total}% >= ${THRESHOLD}%"
