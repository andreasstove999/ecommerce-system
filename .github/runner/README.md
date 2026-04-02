# Self-hosted GitHub Actions runner (Docker)

This directory contains a minimal, single-runner Docker setup for this repository.
It is intended for GitHub Free usage where CI runs on your own machine instead of paid GitHub-hosted runners. It can also be used in other environments where you want to run CI jobs on your own infrastructure.

## What this does

- Builds a container image that installs the official GitHub Actions runner tarball.
- Registers a repository-level self-hosted runner on startup.
- Starts the runner process with labels: `self-hosted,linux,x64,ecommerce`.
- Attempts to de-register the runner on container shutdown.
- Mounts Docker socket (`/var/run/docker.sock`) so future Docker-based jobs are possible.

## Prerequisites

- Docker + Docker Compose plugin.
- Access to repository settings on GitHub.

## Authentication

Use a GitHub Personal Access Token so the container can request short-lived registration/remove tokens automatically.

Recommended minimum permissions:
- Fine-grained PAT: repository access to this repo with **Administration: Read and write**.
- Classic PAT fallback: `repo` scope (broad; use only if needed).

## Setup

1. Copy env file and edit values:

```bash
cd .github/runner
cp .env.example .env
```

2. Update `.env`:
- `GITHUB_OWNER`
- `GITHUB_REPOSITORY`
- `GITHUB_RUNNER_NAME`
- `GITHUB_PAT`

3. Start runner:

```bash
docker compose up -d --build
```

## Verify runner is online

In GitHub:

- Go to **Repository Settings → Actions → Runners**.
- Confirm your runner appears as **Idle/Online** with labels including `ecommerce`.

You can also check local logs:

```bash
docker compose logs -f github-runner
```

## Stop / remove

```bash
docker compose down
```

The entrypoint attempts to remove the runner registration when the container exits.
If a stale offline runner remains in GitHub UI, you can remove it manually from **Settings → Actions → Runners**.

## Offline behavior and merge policy

- If this runner is offline, jobs targeting it will queue until it comes online.
- Pull request merges are blocked **only** if this workflow is configured as a required status check in branch protection.
- If you want offline merges to remain possible, do **not** mark `tests-self-hosted` as required.
