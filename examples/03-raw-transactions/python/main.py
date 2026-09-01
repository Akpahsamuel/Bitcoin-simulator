#!/usr/bin/env python3
"""
Lab 03: Raw Transactions (Python Reference Implementation)
Demonstrates manual UTXO selection, fee calculation, raw transaction construction,
signing, mempool acceptance testing, and broadcast.
"""
import sys
from decimal import Decimal, ROUND_DOWN
from pathlib import Path

# Ensure examples/python is in sys.path
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent.parent
PYTHON_ROOT = REPO_ROOT / "examples" / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

from common.rpc import BitcoinRPC
from common.bootstrap import bootstrap_lab


def main():
    print("=== Lab 03: Raw Transactions (Python) ===")

    try:
        # Step 1: Bootstrap lab
        print("[Step 1] Bootstrapping lab wallet & ensure spendable UTXOs ... ", end="", flush=True)
        node_rpc, wallet_rpc = bootstrap_lab()
        print("✓")

        # Step 2: Select UTXO from listunspent
        print("[Step 2] Selecting UTXO from listunspent ... ", end="", flush=True)
        unspent = wallet_rpc.call("listunspent", [1, 9999999])
        if not unspent:
            raise RuntimeError("No spendable UTXOs found in wallet")

        # Pick the largest UTXO — the shared `lab` wallet also holds small change
        # outputs from earlier lab runs, so the first entry is not reliably big enough.
        utxo = max(unspent, key=lambda u: u["amount"])
        utxo_txid = utxo["txid"]
        utxo_vout = utxo["vout"]
        utxo_amount = Decimal(str(utxo["amount"]))
        print(f"✓ (txid: {utxo_txid[:16]}..., vout: {utxo_vout}, amount: {utxo_amount} BTC)")

        # Step 3: Construct raw transaction
        print("[Step 3] Constructing raw transaction hex (createrawtransaction) ... ", end="", flush=True)
        recipient_addr = wallet_rpc.call("getnewaddress", ["recipient", "bech32m"])
        change_addr = wallet_rpc.call("getnewaddress", ["change", "bech32m"])

        send_amount = Decimal("1.5")
        fee = Decimal("0.0001")
        change_amount = utxo_amount - send_amount - fee
        if change_amount <= 0:
            raise RuntimeError(f"UTXO amount ({utxo_amount}) too small for send + fee")

        inputs = [{"txid": utxo_txid, "vout": utxo_vout}]
        outputs = [
            {recipient_addr: float(send_amount)},
            {change_addr: float(change_amount)},
        ]

        raw_tx_hex = wallet_rpc.call("createrawtransaction", [inputs, outputs])
        assert isinstance(raw_tx_hex, str) and len(raw_tx_hex) > 0, "Invalid raw transaction hex"
        print(f"✓ (hex length: {len(raw_tx_hex)})")

        # Step 4: Sign transaction inputs
        print("[Step 4] Signing transaction inputs (signrawtransactionwithwallet) ... ", end="", flush=True)
        sign_result = wallet_rpc.call("signrawtransactionwithwallet", [raw_tx_hex])
        assert sign_result.get("complete") is True, "Signing incomplete"
        signed_hex = sign_result["hex"]
        print("✓ (complete: true)")

        # Step 5: Verify transaction with testmempoolaccept
        print("[Step 5] Verifying transaction with testmempoolaccept ... ", end="", flush=True)
        mempool_test = node_rpc.call("testmempoolaccept", [[signed_hex]])
        assert mempool_test and mempool_test[0].get("allowed") is True, f"Mempool reject: {mempool_test}"
        print("✓ (allowed: true)")

        # Step 6: Broadcast transaction via sendrawtransaction
        print("[Step 6] Broadcasting transaction via sendrawtransaction ... ", end="", flush=True)
        broadcast_txid = node_rpc.call("sendrawtransaction", [signed_hex])
        assert isinstance(broadcast_txid, str) and len(broadcast_txid) == 64, "Invalid broadcast txid"
        print(f"✓ (txid: {broadcast_txid})")

        # Step 7: Mine 1 block and verify confirmation
        print("[Step 7] Mining 1 block & verifying confirmation ... ", end="", flush=True)
        miner_addr = wallet_rpc.call("getnewaddress", ["miner", "bech32m"])
        wallet_rpc.call("generatetoaddress", [1, miner_addr])

        tx_details = wallet_rpc.call("getrawtransaction", [broadcast_txid, True])
        confs = tx_details.get("confirmations", 0)
        assert confs >= 1, f"Expected confirmations >= 1, got {confs}"
        print(f"✓ (confirmations: {confs})")

        print("======================================================")
        print("Result: PASS ✓")
        print("======================================================")
        sys.exit(0)

    except Exception as exc:
        print(f"\n✗ FAILURE: {exc}")
        print("======================================================")
        print("Result: FAIL ✗")
        print("======================================================")
        sys.exit(1)


if __name__ == "__main__":
    main()
