#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_URL:=https://github.com}"
: "${GITHUB_OWNER:?GITHUB_OWNER is required}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${GITHUB_RUNNER_NAME:=ecommerce-self-hosted}"
: "${GITHUB_RUNNER_LABELS:=self-hosted,linux,x64,ecommerce}"
: "${GITHUB_RUNNER_WORKDIR:=_work}"

REPO_URL="${GITHUB_URL}/${GITHUB_OWNER}/${GITHUB_REPOSITORY}"
API_URL="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPOSITORY}"

if [[ -n "${GITHUB_PAT:-}" ]]; then
  echo "Using GITHUB_PAT to request a short-lived registration token..."
  REGISTRATION_TOKEN="$(curl -fsSL -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_PAT}" \
    "${API_URL}/actions/runners/registration-token" | jq -r '.token')"
elif [[ -n "${REGISTRATION_TOKEN:-}" ]]; then
  echo "Using pre-created REGISTRATION_TOKEN from environment..."
  REGISTRATION_TOKEN="${REGISTRATION_TOKEN}"
else
  echo "Either GITHUB_PAT or REGISTRATION_TOKEN must be set."
  exit 1
fi

if [[ -z "${REGISTRATION_TOKEN}" || "${REGISTRATION_TOKEN}" == "null" ]]; then
  echo "Failed to resolve a runner registration token."
  exit 1
fi

cleanup() {
  echo "Removing runner registration..."
  local remove_token=""

  if [[ -n "${GITHUB_PAT:-}" ]]; then
    remove_token="$(curl -fsSL -X POST \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: Bearer ${GITHUB_PAT}" \
      "${API_URL}/actions/runners/remove-token" | jq -r '.token')"
  fi

  if [[ -n "${remove_token}" && "${remove_token}" != "null" ]]; then
    ./config.sh remove --unattended --token "${remove_token}" || true
  else
    ./config.sh remove --unattended --token "${REGISTRATION_TOKEN}" || true
  fi
}

trap 'cleanup' EXIT INT TERM

./config.sh \
  --url "${REPO_URL}" \
  --token "${REGISTRATION_TOKEN}" \
  --name "${GITHUB_RUNNER_NAME}" \
  --labels "${GITHUB_RUNNER_LABELS}" \
  --work "${GITHUB_RUNNER_WORKDIR}" \
  --unattended \
  --replace

./run.sh
