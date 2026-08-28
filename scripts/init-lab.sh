#!/usr/bin/env bash
set -e

LAB_DIR="${BTC_DATADIR:-$HOME/btc-lab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Initializing Bitcoin Core Workshop Lab ==="

# 1. Ensure lab data directory exists
mkdir -p "$LAB_DIR"

# 2. Write Slide 4 bitcoin.conf if not present
CONF_FILE="$LAB_DIR/bitcoin.conf"
if [ ! -f "$CONF_FILE" ]; then
    echo "Creating $CONF_FILE (Slide 4 configuration)..."
    cat > "$CONF_FILE" <<'EOF'
regtest=1
[regtest]
fallbackfee=0.0002
txindex=1
EOF
else
    echo "Configuration $CONF_FILE already present."
fi


# 4. Start bitcoind daemon if not already running
if pgrep -f "bitcoind.*${LAB_DIR}" >/dev/null 2>&1; then
    echo "Bitcoin Core is already running in regtest mode."
else
    echo "Starting Bitcoin Core daemon (Slide 5)..."
    bitcoind -datadir="$LAB_DIR" -daemon
fi

# 5. Wait for RPC readiness
echo "Waiting for Bitcoin Core RPC to be responsive..."
RETRIES=30
while [ $RETRIES -gt 0 ]; do
    if "$SCRIPT_DIR/btc" getblockchaininfo >/dev/null 2>&1; then
        break
    fi
    sleep 0.5
    RETRIES=$((RETRIES - 1))
done

if [ $RETRIES -eq 0 ]; then
    echo "WARNING: Timed out waiting for bitcoind RPC."
else
    echo "Bitcoin Core RPC is active!"
fi

# 6. Print welcoming status banner
echo ""
echo "================================================================="
echo "  Bitcoin Core 31.1 Regtest Workshop Environment Ready!"
echo "================================================================="
echo "  Chain:      regtest (private local network)"
echo "  Datadir:    $LAB_DIR"
echo "  Command:    btc <rpc-command> (e.g. btc getblockchaininfo)"
echo "  Reset lab:  ./scripts/reset-lab.sh"
echo ""
echo "  You are ready to begin at Slide 6 (createwallet 'lab')."
echo "================================================================="
echo ""

