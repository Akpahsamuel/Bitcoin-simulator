#!/usr/bin/env bash
# One command to bring up the workshop lab:
#   data dir + bitcoin.conf (Slide 4) + a running regtest node (Slide 5).
# Safe to run repeatedly. Also runs automatically on Codespace start.
set -e

LAB_DIR="${BTC_DATADIR:-$HOME/btc-lab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BTC="$SCRIPT_DIR/btc"
BITCOIND="$(command -v bitcoind-real || command -v bitcoind)"

echo "=== Bitcoin Core Workshop Lab ==="

# 1. Data directory
mkdir -p "$LAB_DIR"

# 2. bitcoin.conf (Slide 4) - only written if missing
CONF_FILE="$LAB_DIR/bitcoin.conf"
if [ ! -f "$CONF_FILE" ]; then
    echo "Writing $CONF_FILE (Slide 4 config)..."
    cat > "$CONF_FILE" <<'EOF'
regtest=1
[regtest]
fallbackfee=0.0002
txindex=1
EOF
else
    echo "Config already present: $CONF_FILE"
fi

# 3. Start the node. Idempotent: if RPC already answers, leave it alone.
#    -daemonwait blocks until the node is initialised and RPC is up (Core v25+).
if "$BTC" getblockchaininfo >/dev/null 2>&1; then
    echo "Bitcoin Core already running."
else
    echo "Starting Bitcoin Core (regtest)..."
    "$BITCOIND" -datadir="$LAB_DIR" -daemonwait
fi

# 4. Confirm
echo ""
echo "================================================================="
echo "  Bitcoin Core 31.1 regtest lab is ready."
"$BTC" getblockchaininfo | grep -E '"(chain|blocks)"' || true
echo "  Command:  btc <rpc>          e.g.  btc getblockchaininfo"
echo "  Reset:    bash scripts/reset-lab.sh"
echo "================================================================="
echo ""
