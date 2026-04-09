#!/usr/bin/env bash
set -euo pipefail

THRESHOLD="${COVERAGE_THRESHOLD:-1.0}"

mvn -q -Dcoverage.line.minimum="${THRESHOLD}" verify
