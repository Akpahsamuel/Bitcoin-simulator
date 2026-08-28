FROM debian:bookworm-slim

ARG BITCOIN_VERSION=31.1
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    bash \
    procps \
    jq \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Download, verify, and install Bitcoin Core 31.1 binaries (multi-architecture)
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

# Install bitcoind wrapper to gracefully handle Slide 5 re-runs without lock crashes
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

if pgrep -f "bitcoind.*${DATADIR}" >/dev/null 2>&1; then
    echo "Bitcoin Core starting"
    echo "(Notice: bitcoind is already active in the background for $DATADIR)"
    exit 0
fi

exec /usr/local/bin/bitcoind-real "$@"
EOF

# Standalone btc wrapper on PATH (matches the deck alias)
RUN printf '%s\n' '#!/usr/bin/env bash' \
    'exec bitcoin-cli -datadir="${BTC_DATADIR:-$HOME/btc-lab}" "$@"' \
    > /usr/local/bin/btc && chmod 0755 /usr/local/bin/btc

# Create vscode user with passwordless sudo
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m -s /bin/bash $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Ensure /usr/local/bin is on PATH
ENV PATH="/usr/local/bin:${PATH}"

USER $USERNAME
WORKDIR /home/$USERNAME

