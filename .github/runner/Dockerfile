FROM ubuntu:22.04

ARG RUNNER_VERSION=2.326.0
ENV DEBIAN_FRONTEND=noninteractive

# ── system packages ──────────────────────────────────────────────────
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        gnupg \
        jq \
        tar \
        unzip \
        sudo \
        lsb-release \
        software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# ── Docker CLI (talks to the host daemon via the mounted socket) ─────
RUN install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
       | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
       https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
       > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends docker-ce-cli docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# ── runner user ──────────────────────────────────────────────────────
RUN useradd -m -s /bin/bash runner \
    && usermod -aG sudo runner \
    && echo "runner ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers \
    && mkdir -p /opt/actions-runner \
    && chown -R runner:runner /opt/actions-runner

WORKDIR /opt/actions-runner

# ── download & extract runner as the runner user ─────────────────────
USER runner

RUN curl -fsSL -o actions-runner.tar.gz \
      "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz" \
    && tar xzf actions-runner.tar.gz \
    && rm actions-runner.tar.gz

# ── install runner dependencies (needs root) ─────────────────────────
USER root
RUN ./bin/installdependencies.sh

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER runner

ENTRYPOINT ["/entrypoint.sh"]
