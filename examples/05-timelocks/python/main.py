#!/usr/bin/env python3
"""
Lab 05: Timelocks (Python Reference Implementation)
CLTV (absolute, BIP65) and CSV (relative, BIP112) P2WSH outputs: build the
script, fund it, watch testmempoolaccept reject an early spend, advance the
chain, then spend successfully.
"""
import sys
from pathlib import Path

PYTHON_ROOT = Path(__file__).resolve().parents[2] / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

from common.rpc import BitcoinRPC
from common.bootstrap import bootstrap_lab

from embit import ec, script
from embit.transaction import Transaction, TransactionInput, TransactionOutput, SIGHASH
from embit.script import Witness
from embit.networks import NETWORKS

SK_A_HEX = "11" * 32
EXPECT_CSV_SCRIPT = "53b27521034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aaac"
EXPECT_CSV_ADDR = "bcrt1ql739fkda7sf20qkdwgku2j0ppeff4r7vsqasvvxestsqwtvuak3s9rktmg"

NET = NETWORKS["regtest"]
CSV_DELAY = 3
FUND_BTC = 0.2
FEE_SAT = 20_000

OP_CLTV, OP_CSV, OP_DROP, OP_CHECKSIG = 0xB1, 0xB2, 0x75, 0xAC


def cscriptnum(n: int) -> bytes:
    """Minimal CScriptNum encoding for a non-negative integer."""
    if n == 0:
        return b""
    out = bytearray()
    while n:
        out.append(n & 0xFF)
        n >>= 8
    if out[-1] & 0x80:
        out.append(0x00)
    return bytes(out)


def push(data: bytes) -> bytes:
    assert len(data) < 0x4C, "use OP_PUSHDATA for larger payloads"
    return bytes([len(data)]) + data


def spend(node_rpc, wallet_rpc, priv, witness_script, txid, vout, value_sat, *, version, sequence, locktime):
    """Build, sign and serialise a single-input P2WSH spend back to the lab wallet."""
    dest = wallet_rpc.call("getnewaddress", ["timelock_return", "bech32m"])
    out_sat = value_sat - FEE_SAT
    if out_sat <= 0:
        raise RuntimeError("funded amount too small for fee")
    vin = [TransactionInput(bytes.fromhex(txid), vout, sequence=sequence)]
    vout_ = [TransactionOutput(out_sat, script.address_to_scriptpubkey(dest))]
    tx = Transaction(version=version, vin=vin, vout=vout_, locktime=locktime)
    sighash = tx.sighash_segwit(0, witness_script, value_sat, SIGHASH.ALL)
    sig = priv.sign(sighash).serialize() + bytes([SIGHASH.ALL])
    tx.vin[0].witness = Witness([sig, witness_script.data])
    return tx.serialize().hex()


def find_vout(node_rpc, txid, spk_hex):
    tx = node_rpc.call("getrawtransaction", [txid, True])
    o = next((o for o in tx["vout"] if o["scriptPubKey"]["hex"] == spk_hex), None)
    if o is None:
        raise RuntimeError("funding output not found")
    return o["n"], round(o["value"] * 1e8)


def main():
    print("=== Lab 05: Timelocks (Python) ===")
    try:
        print("[Step 1] Bootstrapping lab wallet & funds ... ", end="", flush=True)
        node_rpc, wallet_rpc = bootstrap_lab()
        priv_a = ec.PrivateKey(bytes.fromhex(SK_A_HEX))
        pub_a = priv_a.get_public_key()
        miner = wallet_rpc.call("getnewaddress", ["lab05_miner", "bech32m"])
        print("✓")

        # ---------- CLTV (absolute) ----------
        print("[Step 2] CLTV: building <height> OP_CLTV script ... ", end="", flush=True)
        lock_height = node_rpc.call("getblockcount") + 10
        cltv_data = push(cscriptnum(lock_height)) + bytes([OP_CLTV, OP_DROP]) + push(pub_a.sec()) + bytes([OP_CHECKSIG])
        cltv_script = script.Script(cltv_data)
        cltv_spk = script.p2wsh(cltv_script)
        print(f"✓ (lock at height {lock_height})")

        print("[Step 3] CLTV: deriving P2WSH address & validating with node ... ", end="", flush=True)
        cltv_addr = cltv_spk.address(NET)
        assert node_rpc.call("validateaddress", [cltv_addr]).get("isvalid") is True
        print(f"✓ ({cltv_addr[:14]}…)")

        print(f"[Step 4] CLTV: funding ({FUND_BTC} BTC) & mining 1 block ... ", end="", flush=True)
        cltv_txid = wallet_rpc.call("sendtoaddress", [cltv_addr, FUND_BTC])
        wallet_rpc.call("generatetoaddress", [1, miner])
        cltv_vout, cltv_sat = find_vout(node_rpc, cltv_txid, cltv_spk.data.hex())
        print(f"✓ (vout {cltv_vout})")

        print("[Step 5] CLTV: early spend rejected by testmempoolaccept ... ", end="", flush=True)
        raw_early = spend(node_rpc, wallet_rpc, priv_a, cltv_script, cltv_txid, cltv_vout, cltv_sat,
                          version=2, sequence=0xFFFFFFFE, locktime=lock_height)
        res = node_rpc.call("testmempoolaccept", [[raw_early]])
        assert res and res[0].get("allowed") is False, f"expected rejection, got {res}"
        print(f"✓ (allowed: false — {res[0].get('reject-reason')})")

        print("[Step 6] CLTV: mining to the lock height ... ", end="", flush=True)
        need = lock_height - node_rpc.call("getblockcount") + 1
        if need > 0:
            wallet_rpc.call("generatetoaddress", [need, miner])
        print(f"✓ (height {node_rpc.call('getblockcount')})")

        print("[Step 7] CLTV: spend accepted, broadcast & confirmed ... ", end="", flush=True)
        raw_ok = spend(node_rpc, wallet_rpc, priv_a, cltv_script, cltv_txid, cltv_vout, cltv_sat,
                       version=2, sequence=0xFFFFFFFE, locktime=lock_height)
        acc = node_rpc.call("testmempoolaccept", [[raw_ok]])
        assert acc and acc[0].get("allowed") is True, f"mempool rejected: {acc}"
        spend_txid = node_rpc.call("sendrawtransaction", [raw_ok])
        wallet_rpc.call("generatetoaddress", [1, miner])
        confs = node_rpc.call("getrawtransaction", [spend_txid, True]).get("confirmations", 0)
        assert confs >= 1, f"confirmations {confs}"
        print(f"✓ (confirmations: {confs})")

        # ---------- CSV (relative) ----------
        print("[Step 8] CSV: building <3> OP_CSV script ... ", end="", flush=True)
        csv_data = bytes([0x50 + CSV_DELAY, OP_CSV, OP_DROP]) + push(pub_a.sec()) + bytes([OP_CHECKSIG])
        csv_script = script.Script(csv_data)
        assert csv_script.data.hex() == EXPECT_CSV_SCRIPT, "CSV script mismatch"
        csv_spk = script.p2wsh(csv_script)
        print("✓ (matches canonical)")

        print("[Step 9] CSV: deriving P2WSH address & validating with node ... ", end="", flush=True)
        csv_addr = csv_spk.address(NET)
        assert csv_addr == EXPECT_CSV_ADDR, f"{csv_addr} != {EXPECT_CSV_ADDR}"
        assert node_rpc.call("validateaddress", [csv_addr]).get("isvalid") is True
        print(f"✓ ({csv_addr[:14]}…)")

        print(f"[Step 10] CSV: funding ({FUND_BTC} BTC) & mining 1 block ... ", end="", flush=True)
        csv_txid = wallet_rpc.call("sendtoaddress", [csv_addr, FUND_BTC])
        wallet_rpc.call("generatetoaddress", [1, miner])
        csv_vout, csv_sat = find_vout(node_rpc, csv_txid, csv_spk.data.hex())
        print(f"✓ (vout {csv_vout}, 1 confirmation)")

        print("[Step 11] CSV: early spend rejected by testmempoolaccept ... ", end="", flush=True)
        raw_early = spend(node_rpc, wallet_rpc, priv_a, csv_script, csv_txid, csv_vout, csv_sat,
                          version=2, sequence=CSV_DELAY, locktime=0)
        res = node_rpc.call("testmempoolaccept", [[raw_early]])
        assert res and res[0].get("allowed") is False, f"expected rejection, got {res}"
        print(f"✓ (allowed: false — {res[0].get('reject-reason')})")

        print("[Step 12] CSV: mining 3 blocks to satisfy the relative delay ... ", end="", flush=True)
        wallet_rpc.call("generatetoaddress", [CSV_DELAY, miner])
        print("✓")

        print("[Step 13] CSV: spend accepted, broadcast & confirmed ... ", end="", flush=True)
        raw_ok = spend(node_rpc, wallet_rpc, priv_a, csv_script, csv_txid, csv_vout, csv_sat,
                       version=2, sequence=CSV_DELAY, locktime=0)
        acc = node_rpc.call("testmempoolaccept", [[raw_ok]])
        assert acc and acc[0].get("allowed") is True, f"mempool rejected: {acc}"
        spend_txid = node_rpc.call("sendrawtransaction", [raw_ok])
        wallet_rpc.call("generatetoaddress", [1, miner])
        confs = node_rpc.call("getrawtransaction", [spend_txid, True]).get("confirmations", 0)
        assert confs >= 1, f"confirmations {confs}"
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
