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

# ── Ensure the runner user can access the Docker socket ──────────────
if [ -S /var/run/docker.sock ]; then
  DOCKER_GID="$(stat -c '%g' /var/run/docker.sock)"
  if ! getent group "${DOCKER_GID}" > /dev/null 2>&1; then
    sudo groupadd -g "${DOCKER_GID}" dockerhost
  fi
  DOCKER_GROUP="$(getent group "${DOCKER_GID}" | cut -d: -f1)"
  if ! id -nG runner | grep -qw "${DOCKER_GROUP}"; then
    sudo usermod -aG "${DOCKER_GROUP}" runner
  fi
fi

# ── Ensure the runner owns its work directory (volume may be root) ────
sudo chown -R runner:runner /opt/actions-runner/_work

# ── Obtain a short-lived registration token ──────────────────────────
if [[ -n "${GITHUB_PAT:-}" ]]; then
  echo "Using GITHUB_PAT to request a short-lived registration token..."
  REGISTRATION_TOKEN="$(curl -fsSL -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ${GITHUB_PAT}" \
    "${API_URL}/actions/runners/registration-token" | jq -r '.token')"
else
  echo "GITHUB_PAT must be set."
  exit 1
fi

if [[ -z "${REGISTRATION_TOKEN}" || "${REGISTRATION_TOKEN}" == "null" ]]; then
  echo "Failed to resolve a runner registration token."
  exit 1
fi

# ── Cleanup on exit ──────────────────────────────────────────────────
cleanup() {
  echo "Removing runner registration..."
  local remove_token=""

  if [[ -n "${GITHUB_PAT:-}" ]]; then
    remove_token="$(curl -fsSL -X POST \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: Bearer ${GITHUB_PAT}" \
      "${API_URL}/actions/runners/remove-token" | jq -r '.token')" || true
  fi

  if [[ -n "${remove_token}" && "${remove_token}" != "null" ]]; then
    ./config.sh remove --unattended --token "${remove_token}" || true
  else
    echo "Failed to deregister runner cleanly: could not get remove_token."
  fi
}

trap 'cleanup' EXIT INT TERM

# ── Remove stale configuration from previous container runs ──────────
if [ -f .runner ]; then
  echo "Removing stale runner configuration..."
  ./config.sh remove --unattended --token "${REGISTRATION_TOKEN}" 2>/dev/null || true
fi

# ── Configure the runner ─────────────────────────────────────────────
./config.sh \
  --url "${REPO_URL}" \
  --token "${REGISTRATION_TOKEN}" \
  --name "${GITHUB_RUNNER_NAME}" \
  --labels "${GITHUB_RUNNER_LABELS}" \
  --work "${GITHUB_RUNNER_WORKDIR}" \
  --unattended \
  --replace

# ── Start the runner (with automatic retry on transient errors) ──────
if [ -S /var/run/docker.sock ]; then
  # Use sudo to spawn run.sh as the same user so the new group membership takes effect
  sudo -E -u runner ./run.sh
else
  ./run.sh
fi
