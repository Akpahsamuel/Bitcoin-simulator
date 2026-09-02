#!/usr/bin/env python3
"""
Lab 06: ZMQ Listener (Python Reference Implementation)
Subscribe to Bitcoin Core's rawtx / rawblock ZeroMQ feeds, trigger one
transaction and one block, and verify the pushed bytes against the node.
"""
import sys
from pathlib import Path

PYTHON_ROOT = Path(__file__).resolve().parents[2] / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

import zmq

from common.rpc import BitcoinRPC, get_config
from common.bootstrap import bootstrap_lab

RECV_TIMEOUT_MS = 15_000


def main():
    print("=== Lab 06: ZMQ Listener (Python) ===")
    ctx = zmq.Context()
    try:
        # Step 1
        print("[Step 1] Bootstrapping lab wallet & funds ... ", end="", flush=True)
        node_rpc, wallet_rpc = bootstrap_lab()
        cfg = get_config()
        print("✓")

        # Step 2
        print(
            f"[Step 2] Subscribing to rawtx ({cfg['zmq_rawtx']}) and rawblock ({cfg['zmq_rawblock']}) ... ",
            end="",
            flush=True,
        )
        sub_tx = ctx.socket(zmq.SUB)
        sub_tx.setsockopt(zmq.RCVTIMEO, RECV_TIMEOUT_MS)
        sub_tx.setsockopt(zmq.SUBSCRIBE, b"rawtx")
        sub_tx.connect(cfg["zmq_rawtx"])

        sub_block = ctx.socket(zmq.SUB)
        sub_block.setsockopt(zmq.RCVTIMEO, RECV_TIMEOUT_MS)
        sub_block.setsockopt(zmq.SUBSCRIBE, b"rawblock")
        sub_block.connect(cfg["zmq_rawblock"])

        # let the SUB subscriptions propagate before we trigger events
        import time
        time.sleep(0.5)
        print("✓")

        # Step 3
        print("[Step 3] Broadcasting a transaction ... ", end="", flush=True)
        dest = wallet_rpc.call("getnewaddress", ["zmq_probe", "bech32m"])
        sent_txid = wallet_rpc.call("sendtoaddress", [dest, 0.01])
        print(f"✓ (txid: {sent_txid[:16]}…)")

        # Step 4
        print("[Step 4] Received rawtx frame & verified txid ... ", end="", flush=True)
        try:
            topic, body, seq = sub_tx.recv_multipart()
        except zmq.Again:
            raise RuntimeError(
                f"no rawtx notification within {RECV_TIMEOUT_MS} ms — is zmqpubrawtx set? "
                "(see docs/troubleshooting.md)"
            )
        assert topic == b"rawtx", f"unexpected topic {topic!r}"
        decoded = node_rpc.call("decoderawtransaction", [body.hex()])
        assert decoded["txid"] == sent_txid, f"rawtx txid {decoded['txid']} != {sent_txid}"
        print(f"✓ (seq {int.from_bytes(seq, 'little')})")

        # Step 5
        print("[Step 5] Mining 1 block ... ", end="", flush=True)
        miner = wallet_rpc.call("getnewaddress", ["lab06_miner", "bech32m"])
        block_hash = wallet_rpc.call("generatetoaddress", [1, miner])[0]
        print(f"✓ (hash: {block_hash[:16]}…)")

        # Step 6
        print("[Step 6] Received rawblock frame & verified against getblock ... ", end="", flush=True)
        try:
            topic, body, seq = sub_block.recv_multipart()
        except zmq.Again:
            raise RuntimeError(
                f"no rawblock notification within {RECV_TIMEOUT_MS} ms — is zmqpubrawblock set? "
                "(see docs/troubleshooting.md)"
            )
        assert topic == b"rawblock", f"unexpected topic {topic!r}"
        expected = node_rpc.call("getblock", [block_hash, 0])
        assert body.hex() == expected, "rawblock payload does not match getblock"
        print(f"✓ (seq {int.from_bytes(seq, 'little')})")

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
    finally:
        ctx.destroy(linger=0)


if __name__ == "__main__":
    main()
