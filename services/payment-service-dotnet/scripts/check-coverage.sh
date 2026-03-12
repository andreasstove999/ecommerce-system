#!/usr/bin/env bash
set -euo pipefail

THRESHOLD="${COVERAGE_THRESHOLD:-100}"

DOTNET_CLI_TELEMETRY_OPTOUT=1 \
  dotnet test tests/PaymentService.Tests/PaymentService.Tests.csproj \
  /p:CollectCoverage=true \
  /p:CoverletOutputFormat=cobertura \
  /p:Threshold="${THRESHOLD}" \
  /p:ThresholdType=line \
  /p:ThresholdStat=total
