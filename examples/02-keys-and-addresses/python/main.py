#!/usr/bin/env python3
"""
Lab 02: Keys and Addresses (Python Reference Implementation)
Demonstrates BIP39 mnemonics, BIP32/BIP44 derivation, Legacy/SegWit/Taproot addresses,
and node consensus validation.
"""
import sys
from pathlib import Path

# Ensure examples/python is in sys.path
SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent.parent
PYTHON_ROOT = REPO_ROOT / "examples" / "python"
if str(PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(PYTHON_ROOT))

from common.rpc import BitcoinRPC
from common.bootstrap import bootstrap_lab

from embit import bip39, bip32, script
from embit.networks import NETWORKS

# Deterministic test vector: the canonical "abandon ... about" mnemonic on regtest.
# Expected values are cross-checked against bitcoinjs-lib (TypeScript), rust-bitcoin
# (Rust) and btcd (Go) — every port must derive exactly these.
TEST_MNEMONIC = (
    "abandon abandon abandon abandon abandon abandon abandon abandon "
    "abandon abandon abandon about"
)
EXPECT_FINGERPRINT = "73c5da0a"
EXPECT_P2PKH = "mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV"
EXPECT_P2WPKH = "bcrt1q6rz28mcfaxtmd6v789l9rrlrusdprr9pz3cppk"
EXPECT_P2TR = "bcrt1p8wpt9v4frpf3tkn0srd97pksgsxc5hs52lafxwru9kgeephvs7rqjeprhg"


def main():
    print("=== Lab 02: Keys & Addresses (Python) ===")

    try:
        # Step 1: Bootstrap lab
        print("[Step 1] Bootstrapping lab wallet & RPC connection ... ", end="", flush=True)
        node_rpc, _ = bootstrap_lab()
        print("✓")

        # Step 2: BIP39 mnemonic -> 512-bit root seed
        print("[Step 2] Generating BIP39 mnemonic & root seed ... ", end="", flush=True)
        seed = bip39.mnemonic_to_seed(TEST_MNEMONIC)
        assert len(seed) == 64, f"Expected 64-byte seed, got {len(seed)}"
        print(f"✓ (12 words, {len(seed)}-byte seed)")

        # Step 3: BIP32 root key
        print("[Step 3] Deriving BIP32 root key ... ", end="", flush=True)
        root = bip32.HDKey.from_seed(seed, version=NETWORKS["regtest"]["xprv"])
        fingerprint = root.my_fingerprint.hex()
        assert fingerprint == EXPECT_FINGERPRINT, f"fingerprint {fingerprint} != {EXPECT_FINGERPRINT}"
        print(f"✓ (fingerprint: {fingerprint})")

        # Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
        print("[Step 4] Deriving BIP44 Legacy P2PKH address ... ", end="", flush=True)
        k_p2pkh = root.derive("m/44h/1h/0h/0/0").key
        addr_p2pkh = script.p2pkh(k_p2pkh).address(NETWORKS["regtest"])
        assert addr_p2pkh == EXPECT_P2PKH, f"P2PKH {addr_p2pkh} != {EXPECT_P2PKH}"
        print(f"✓ ({addr_p2pkh})")

        # Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
        print("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ", end="", flush=True)
        k_p2wpkh = root.derive("m/84h/1h/0h/0/0").key
        addr_p2wpkh = script.p2wpkh(k_p2wpkh).address(NETWORKS["regtest"])
        assert addr_p2wpkh == EXPECT_P2WPKH, f"P2WPKH {addr_p2wpkh} != {EXPECT_P2WPKH}"
        print(f"✓ ({addr_p2wpkh})")

        # Step 6: Taproot P2TR (m/86'/1'/0'/0/0, BIP86 key-path tweak)
        print("[Step 6] Deriving BIP86 Taproot P2TR address ... ", end="", flush=True)
        k_p2tr = root.derive("m/86h/1h/0h/0/0").key
        addr_p2tr = script.p2tr(k_p2tr).address(NETWORKS["regtest"])
        assert addr_p2tr == EXPECT_P2TR, f"P2TR {addr_p2tr} != {EXPECT_P2TR}"
        print(f"✓ ({addr_p2tr})")

        # Step 7: Validate derived addresses with Bitcoin Core node
        print("[Step 7] Validating addresses against Bitcoin Core node ... ", end="", flush=True)
        for addr in [addr_p2pkh, addr_p2wpkh, addr_p2tr]:
            val = node_rpc.call("validateaddress", [addr])
            assert val.get("isvalid") is True, f"Address {addr} reported invalid by node!"
        print("✓ (all addresses valid)")

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
