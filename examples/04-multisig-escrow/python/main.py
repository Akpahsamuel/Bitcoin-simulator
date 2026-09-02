#!/usr/bin/env python3
"""
Lab 04: Multisig Escrow (Python Reference Implementation)
2-of-3 P2WSH built from raw script, funded from the lab wallet, then spent by
assembling the witness stack [<empty>, sigA, sigB, witnessScript] by hand.
"""
import sys
from pathlib import Path

# .../examples/<NN>/python/main.py → parents[2] == .../examples → examples/python holds common/
PYTHON_ROOT = Path(__file__).resolve().parents[2] / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

from common.rpc import BitcoinRPC
from common.bootstrap import bootstrap_lab

from embit import ec, script
from embit.transaction import Transaction, TransactionInput, TransactionOutput, SIGHASH
from embit.script import Witness
from embit.networks import NETWORKS

# Fixed escrow keys — every language port derives the same script/address.
SK_HEX = [
    "11" * 32,  # key A
    "22" * 32,  # key B
    "33" * 32,  # key C
]
EXPECT_WITNESS_SCRIPT = (
    "5221034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa"
    "2102466d7fcae563e5cb09a0d1870bb580344804617879a14949cf22285f1bae3f27"
    "21023c72addb4fdf09af94f0c94d7fe92a386a7e70cf8a1d85916386bb2535c7b1b153ae"
)
EXPECT_P2WSH = "bcrt1qpy8yjjs2l5neewx722mxve9w6m77zqsu7rldukggseflhwralerqh6ma0d"

NET = NETWORKS["regtest"]
FUND_BTC = 0.5
FEE_SAT = 20_000


def main():
    print("=== Lab 04: Multisig Escrow (Python) ===")
    try:
        # Step 1: Bootstrap
        print("[Step 1] Bootstrapping lab wallet & funds ... ", end="", flush=True)
        node_rpc, wallet_rpc = bootstrap_lab()
        print("✓")

        # Step 2: Derive 3 escrow keypairs
        print("[Step 2] Deriving 3 escrow keypairs from test vectors ... ", end="", flush=True)
        privs = [ec.PrivateKey(bytes.fromhex(h)) for h in SK_HEX]
        pubs = [p.get_public_key() for p in privs]
        print(f"✓ ({', '.join(p.sec().hex()[:10] + '…' for p in pubs)})")

        # Step 3: 2-of-3 witness script
        print("[Step 3] Building 2-of-3 witness script ... ", end="", flush=True)
        witness_script = script.multisig(2, pubs)
        assert witness_script.data.hex() == EXPECT_WITNESS_SCRIPT, "witness script mismatch"
        print("✓ (matches canonical)")

        # Step 4: P2WSH address
        print("[Step 4] Deriving P2WSH address & validating with node ... ", end="", flush=True)
        spk = script.p2wsh(witness_script)
        addr = spk.address(NET)
        assert addr == EXPECT_P2WSH, f"address {addr} != {EXPECT_P2WSH}"
        val = node_rpc.call("validateaddress", [addr])
        assert val.get("isvalid") is True, "node rejected the P2WSH address"
        print(f"✓ ({addr[:14]}…)")

        # Step 5: Fund the multisig
        print(f"[Step 5] Funding the multisig ({FUND_BTC} BTC) & mining 1 block ... ", end="", flush=True)
        fund_txid = wallet_rpc.call("sendtoaddress", [addr, FUND_BTC])
        miner = wallet_rpc.call("getnewaddress", ["lab04_miner", "bech32m"])
        wallet_rpc.call("generatetoaddress", [1, miner])
        print(f"✓ (txid: {fund_txid[:16]}…)")

        # Step 6: Locate the funding UTXO by scriptPubKey
        print("[Step 6] Locating the funding UTXO by scriptPubKey ... ", end="", flush=True)
        fund_tx = node_rpc.call("getrawtransaction", [fund_txid, True])
        spk_hex = spk.data.hex()
        vout = next((o for o in fund_tx["vout"] if o["scriptPubKey"]["hex"] == spk_hex), None)
        if vout is None:
            raise RuntimeError("funding output not found in transaction")
        fund_vout = vout["n"]
        fund_sat = round(vout["value"] * 1e8)
        print(f"✓ (vout: {fund_vout}, {vout['value']} BTC)")

        # Step 7: Build the spend transaction
        print("[Step 7] Building the spend transaction ... ", end="", flush=True)
        return_addr = wallet_rpc.call("getnewaddress", ["escrow_return", "bech32m"])
        out_sat = fund_sat - FEE_SAT
        if out_sat <= 0:
            raise RuntimeError("funded amount too small for fee")
        vin = [TransactionInput(bytes.fromhex(fund_txid), fund_vout)]
        vout_ = [TransactionOutput(out_sat, script.address_to_scriptpubkey(return_addr))]
        tx = Transaction(version=2, vin=vin, vout=vout_, locktime=0)
        print("✓")

        # Step 8: BIP143 sighash + sign with keys A and B
        print("[Step 8] BIP143 sighash + signing with keys A and B ... ", end="", flush=True)
        sighash = tx.sighash_segwit(0, witness_script, fund_sat, SIGHASH.ALL)
        sig_a = privs[0].sign(sighash).serialize() + bytes([SIGHASH.ALL])
        sig_b = privs[1].sign(sighash).serialize() + bytes([SIGHASH.ALL])
        print(f"✓ (sigA {len(sig_a)}B, sigB {len(sig_b)}B)")

        # Step 9: Assemble the witness stack
        print("[Step 9] Assembling witness [<empty>, sigA, sigB, script] ... ", end="", flush=True)
        tx.vin[0].witness = Witness([b"", sig_a, sig_b, witness_script.data])
        raw_hex = tx.serialize().hex()
        print(f"✓ (tx {len(raw_hex)//2} bytes)")

        # Step 10: testmempoolaccept
        print("[Step 10] testmempoolaccept ... ", end="", flush=True)
        accept = node_rpc.call("testmempoolaccept", [[raw_hex]])
        if not (accept and accept[0].get("allowed") is True):
            raise RuntimeError(f"mempool rejected: {accept}")
        print("✓ (allowed: true)")

        # Step 11: Broadcast & confirm
        print("[Step 11] Broadcasting & mining 1 block ... ", end="", flush=True)
        spend_txid = node_rpc.call("sendrawtransaction", [raw_hex])
        wallet_rpc.call("generatetoaddress", [1, miner])
        details = node_rpc.call("getrawtransaction", [spend_txid, True])
        confs = details.get("confirmations", 0)
        assert confs >= 1, f"expected confirmations >= 1, got {confs}"
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
