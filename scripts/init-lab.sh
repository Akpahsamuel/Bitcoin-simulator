#!/usr/bin/env bash
# Initialize and launch the Bitcoin Developer Sandbox & Starter Kit regtest node:
#   1. Sets up data directory.
#   2. Emits full bitcoin.conf (RPC, REST, ZMQ enabled).
#   3. Generates root .env from .env.example if missing.
#   4. Starts the regtest node at height 0 (idempotent, safe to re-run).
set -e

LAB_DIR="${BTC_DATADIR:-$HOME/btc-lab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BTC="$SCRIPT_DIR/btc"
BITCOIND="$(command -v bitcoind-real || command -v bitcoind || true)"

echo "=== Bitcoin Developer Sandbox: Lab Init ==="

# 1. Data directory
mkdir -p "$LAB_DIR"

# 2. bitcoin.conf - write if missing, or refresh if a previous (older) init
#    left a config without the REST/ZMQ lines this sandbox now needs.
CONF_FILE="$LAB_DIR/bitcoin.conf"
write_conf() {
    cat > "$CONF_FILE" <<'CONF'
regtest=1
server=1
rest=1
txindex=1
rpcuser=bitcoinrpc
rpcpassword=bitcoinrpcpassword
rpcbind=0.0.0.0
rpcallowip=0.0.0.0/0
[regtest]
fallbackfee=0.0002
zmqpubrawblock=tcp://0.0.0.0:28332
zmqpubrawtx=tcp://0.0.0.0:28333
CONF
}
if [ ! -f "$CONF_FILE" ]; then
    echo "Writing $CONF_FILE (RPC, REST, ZMQ enabled)..."
    write_conf
elif ! grep -q '^zmqpubrawblock=' "$CONF_FILE"; then
    echo "Existing $CONF_FILE is missing REST/ZMQ settings - refreshing (old copy -> bitcoin.conf.bak)..."
    cp "$CONF_FILE" "$CONF_FILE.bak"
    write_conf
else
    echo "Config already present: $CONF_FILE"
fi

# 3. .env file from .env.example
ENV_FILE="$REPO_ROOT/.env"
ENV_EXAMPLE="$REPO_ROOT/.env.example"
if [ ! -f "$ENV_FILE" ]; then
    if [ -f "$ENV_EXAMPLE" ]; then
        echo "Creating .env from .env.example..."
        cp "$ENV_EXAMPLE" "$ENV_FILE"
    fi
else
    echo "Environment file already present: .env"
fi

# 4. Start the node. Idempotent: if RPC already answers, leave it alone.
#    -daemonwait blocks until the node is initialized and RPC is up (Core v25+).
if [ -n "$BITCOIND" ]; then
    if "$BTC" getblockchaininfo >/dev/null 2>&1; then
        echo "Bitcoin Core is already running."
    else
        echo "Starting Bitcoin Core (regtest)..."
        "$BITCOIND" -datadir="$LAB_DIR" -daemonwait
    fi

    echo ""
    echo "================================================================="
    echo "  Bitcoin Core 31.1 regtest sandbox is ready."
    "$BTC" getblockchaininfo | grep -E '"(chain|blocks)"' || true
    echo "  Endpoints:"
    echo "    - JSON-RPC: http://127.0.0.1:18443 (user: bitcoinrpc)"
    echo "    - REST API: http://127.0.0.1:18443/rest/"
    echo "    - ZMQ Block: tcp://127.0.0.1:28332"
    echo "    - ZMQ Tx:    tcp://127.0.0.1:28333"
    echo "  Commands:"
    echo "    - CLI:      btc <rpc>          e.g. btc getblockchaininfo"
    echo "    - Reset:    bash scripts/reset-lab.sh"
    echo "================================================================="
    echo ""
else
    echo "bitcoind binary not found on PATH. Config and .env prepared."
    echo "Inside a devcontainer or after installing Bitcoin Core, run: bash scripts/init-lab.sh"
fi
