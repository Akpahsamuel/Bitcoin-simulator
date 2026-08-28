#!/usr/bin/env bash
set -e

LAB_DIR="${BTC_DATADIR:-$HOME/btc-lab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BTC="$SCRIPT_DIR/btc"

echo "=== Resetting Bitcoin Core Workshop Lab (Slide 16) ==="

# 1. Stop node
echo "Stopping Bitcoin Core..."
"$BTC" stop >/dev/null 2>&1 || true

# Wait for process to stop
RETRIES=20
while pgrep -f "bitcoind.*${LAB_DIR}" >/dev/null 2>&1 && [ $RETRIES -gt 0 ]; do
    sleep 0.5
    RETRIES=$((RETRIES - 1))
done

# 2. Wipe regtest data (keep bitcoin.conf)
echo "Wiping regtest data directory ($LAB_DIR/regtest)..."
rm -rf "$LAB_DIR/regtest"

# 3. Start clean daemon
echo "Restarting bitcoind on fresh regtest chain..."
bitcoind -datadir="$LAB_DIR" -daemon

# 4. Wait for RPC readiness
echo "Waiting for Bitcoin Core RPC..."
RETRIES=30
while [ $RETRIES -gt 0 ]; do
    if "$BTC" getblockchaininfo >/dev/null 2>&1; then
        break
    fi
    sleep 0.5
    RETRIES=$((RETRIES - 1))
done

echo ""
echo "=== Lab successfully reset! ==="
"$BTC" getblockchaininfo | grep -E '"(chain|blocks)"' || true
echo ""

