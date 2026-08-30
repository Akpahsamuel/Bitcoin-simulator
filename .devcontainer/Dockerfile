FROM debian:bookworm-slim

ARG BITCOIN_VERSION=31.1
ARG GO_VERSION=1.22.12
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

ENV DEBIAN_FRONTEND=noninteractive

# 1. Install system utilities and build prerequisites
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    bash \
    procps \
    jq \
    sudo \
    build-essential \
    pkg-config \
    libzmq3-dev \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# 2. Install Node.js 20 LTS
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# 3. Install Go 1.22
RUN set -ex; \
    ARCH="$(uname -m)"; \
    case "${ARCH}" in \
        x86_64) GO_ARCH="amd64" ;; \
        aarch64|arm64) GO_ARCH="arm64" ;; \
        *) echo "Unsupported architecture for Go: ${ARCH}" && exit 1 ;; \
    esac; \
    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz; \
    tar -C /usr/local -xzf /tmp/go.tar.gz; \
    rm -f /tmp/go.tar.gz

# 4. Download, verify, and install Bitcoin Core 31.1 binaries (multi-architecture)
RUN set -ex; \
    ARCH="$(uname -m)"; \
    case "${ARCH}" in \
        x86_64) BITCOIN_ARCH="x86_64-linux-gnu" ;; \
        aarch64|arm64) BITCOIN_ARCH="aarch64-linux-gnu" ;; \
        *) echo "Unsupported architecture: ${ARCH}" && exit 1 ;; \
    esac; \
    BITCOIN_TAR="bitcoin-${BITCOIN_VERSION}-${BITCOIN_ARCH}.tar.gz"; \
    BASE_URL="https://bitcoincore.org/bin/bitcoin-core-${BITCOIN_VERSION}"; \
    cd /tmp; \
    curl -fsSL "${BASE_URL}/${BITCOIN_TAR}" -o "${BITCOIN_TAR}"; \
    curl -fsSL "${BASE_URL}/SHA256SUMS" -o SHA256SUMS; \
    grep " ${BITCOIN_TAR}\$" SHA256SUMS | sha256sum -c -; \
    mkdir -p /tmp/bitcoin; \
    tar -xzf "${BITCOIN_TAR}" -C /tmp/bitcoin --strip-components=1; \
    install -m 0755 /tmp/bitcoin/bin/bitcoind     /usr/local/bin/bitcoind-real; \
    install -m 0755 /tmp/bitcoin/bin/bitcoin-cli  /usr/local/bin/bitcoin-cli; \
    install -m 0755 /tmp/bitcoin/bin/bitcoin-tx   /usr/local/bin/bitcoin-tx; \
    install -m 0755 /tmp/bitcoin/bin/bitcoin-util /usr/local/bin/bitcoin-util; \
    rm -rf /tmp/bitcoin*

# 5. Install bitcoind wrapper to gracefully handle re-runs without lock crashes
RUN cat <<'EOF' > /usr/local/bin/bitcoind && chmod +x /usr/local/bin/bitcoind
#!/usr/bin/env bash
for arg in "$@"; do
    case "$arg" in
        --help|-help|-h|-version|--version)
            exec /usr/local/bin/bitcoind-real "$@"
            ;;
    esac
done

DATADIR="${BTC_DATADIR:-$HOME/btc-lab}"
for arg in "$@"; do
    case "$arg" in
        -datadir=*) DATADIR="${arg#-datadir=}" ;;
    esac
done

if pgrep -x bitcoind-real >/dev/null 2>&1; then
    echo "Bitcoin Core is already running for $DATADIR - leaving it running."
    exit 0
fi

exec /usr/local/bin/bitcoind-real "$@"
EOF

# 6. Standalone btc wrapper on PATH
RUN printf '%s\n' '#!/usr/bin/env bash' \
    'exec bitcoin-cli -datadir="${BTC_DATADIR:-$HOME/btc-lab}" "$@"' \
    > /usr/local/bin/btc && chmod 0755 /usr/local/bin/btc

# 7. Create vscode user with passwordless sudo
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m -s /bin/bash $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# 8. Install Rust for vscode user via rustup
USER $USERNAME
WORKDIR /home/$USERNAME

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable

# 9. Configure PATH environment variables
ENV PATH="/usr/local/go/bin:/home/${USERNAME}/go/bin:/home/${USERNAME}/.cargo/bin:/usr/local/bin:${PATH}"
ENV GOPATH="/home/${USERNAME}/go"

CMD ["/bin/bash"]
