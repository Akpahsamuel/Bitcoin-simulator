#!/usr/bin/env bash
# Wipe the regtest chain and restart on a fresh one. Keeps bitcoin.conf and .env.
set -e

LAB_DIR="${BTC_DATADIR:-$HOME/btc-lab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BTC="$SCRIPT_DIR/btc"
BITCOIND="$(command -v bitcoind-real || command -v bitcoind || true)"

echo "=== Resetting Bitcoin Developer Sandbox ==="

# 1. Stop the node and wait for it to exit
if [ -n "$BITCOIND" ]; then
    echo "Stopping Bitcoin Core..."
    "$BTC" stop >/dev/null 2>&1 || true
    RETRIES=20
    while pgrep -x bitcoind-real >/dev/null 2>&1 && [ $RETRIES -gt 0 ]; do
        sleep 0.5
        RETRIES=$((RETRIES - 1))
    done

    # 2. Wipe regtest data (keep bitcoin.conf)
    echo "Wiping $LAB_DIR/regtest ..."
    rm -rf "$LAB_DIR/regtest"

    # 3. Start a clean node (blocks until RPC is ready)
    echo "Starting fresh node..."
    "$BITCOIND" -datadir="$LAB_DIR" -daemonwait

    echo ""
    echo "=== Sandbox reset complete. ==="
    "$BTC" getblockchaininfo | grep -E '"(chain|blocks)"' || true
    echo ""
else
    echo "bitcoind binary not found on PATH."
    if [ -d "$LAB_DIR/regtest" ]; then
        echo "Wiping $LAB_DIR/regtest ..."
        rm -rf "$LAB_DIR/regtest"
        echo "Cleaned regtest directory."
    fi
fi
