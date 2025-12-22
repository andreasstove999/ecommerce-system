#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTRACTS_DIR="${ROOT_DIR}/contracts"

if [[ ! -d "${CONTRACTS_DIR}" ]]; then
  echo "Contracts directory not found at ${CONTRACTS_DIR}" >&2
  exit 1
fi

npm --prefix "${CONTRACTS_DIR}" install --no-fund --no-audit
npm --prefix "${CONTRACTS_DIR}" run validate
