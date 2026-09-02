#!/usr/bin/env python3
"""
Lab 01: RPC Client (Python Reference Implementation)
Demonstrates JSON-RPC authentication, wallet operations, mining, and REST queries.
"""
import sys
from pathlib import Path

# .../examples/<NN>/python/main.py -> parents[2] == .../examples -> examples/python holds common/
PYTHON_ROOT = Path(__file__).resolve().parents[2] / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

from common.rpc import BitcoinRPC
from common.bootstrap import bootstrap_lab


def main():
    print("=== Lab 01: RPC Client (Python) ===")

    try:
        # Step 1: Bootstrap lab wallet and coinbase maturity
        print("[Step 1] Bootstrapping lab wallet and initial funds ... ", end="", flush=True)
        node_rpc, wallet_rpc = bootstrap_lab()
        print("✓")

        # Step 2: Query blockchain info via JSON-RPC
        print("[Step 2] Querying getblockchaininfo via JSON-RPC ... ", end="", flush=True)
        chain_info = node_rpc.call("getblockchaininfo")
        assert chain_info["chain"] == "regtest", f"Expected regtest, got {chain_info['chain']}"
        assert chain_info["blocks"] >= 101, f"Expected at least 101 blocks, got {chain_info['blocks']}"
        print(f"✓ (chain: {chain_info['chain']}, blocks: {chain_info['blocks']})")

        # Step 3: Generate fresh address and mine 1 block
        print("[Step 3] Generating fresh address and mining 1 block ... ", end="", flush=True)
        fresh_addr = wallet_rpc.call("getnewaddress", ["lab01_test", "bech32m"])
        block_hashes = wallet_rpc.call("generatetoaddress", [1, fresh_addr])
        assert len(block_hashes) == 1, "Expected 1 block mined"
        new_blocks = node_rpc.call("getblockcount")
        print(f"✓ (new height: {new_blocks}, mined block: {block_hashes[0][:16]}...)")

        # Step 4: Query wallet balance
        print("[Step 4] Querying wallet balance via getbalance ... ", end="", flush=True)
        balance = wallet_rpc.call("getbalance")
        assert balance > 0, f"Expected spendable balance > 0, got {balance}"
        print(f"✓ (balance: {balance} BTC)")

        # Step 5: Query unauthenticated REST API
        print("[Step 5] Querying unauthenticated REST API (/rest/chaininfo.json) ... ", end="", flush=True)
        rest_info = node_rpc.get_rest("chaininfo.json")
        assert rest_info["chain"] == "regtest", "REST returned unexpected chain"
        assert rest_info["blocks"] == new_blocks, "REST block count mismatch"
        print("✓")

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
