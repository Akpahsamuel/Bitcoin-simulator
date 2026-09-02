# 02 · Keys & Addresses (Cryptographic Primitives & Derivation)

This lab explores cryptographic key derivation in Bitcoin — from a BIP39 mnemonic seed phrase to BIP32/BIP44 Hierarchical Deterministic (HD) wallets and modern address encodings (Legacy P2PKH, Native SegWit BIP84 P2WPKH, and Taproot BIP86 P2TR).

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**:
  - [Chapter 4: Keys and Addresses](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch04_keys.adoc)
  - [Chapter 5: Wallets](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch05_wallets.adoc)
- **Mastering the Lightning Network**:
  - [Chapter 6: Lightning Architecture](https://github.com/lnbook/lnbook/blob/develop/06_lightning_architecture.asciidoc)

---

## Concepts Demonstrated
1. **BIP39 Mnemonic & Seed Generation**: Converting entropy into 12/24-word human-readable mnemonics and generating a 512-bit binary root seed.
2. **BIP32 Hierarchical Deterministic Derivation**: Deriving master extended private keys (`xprv`/`tprv`), master public keys (`xpub`/`tpub`), and child keys.
3. **Address Standards**:
   - **Legacy P2PKH** (BIP44 path `m/44'/1'/0'/0/0`): Hash160 public key hash (`m...` / `n...` on regtest).
   - **Native SegWit P2WPKH** (BIP84 path `m/84'/1'/0'/0/0`): Witness v0 Bech32 (`bcrt1q...` on regtest).
   - **Taproot P2TR** (BIP86 path `m/86'/1'/0'/0/0`): Witness v1 Bech32m with Schnorr output key (`bcrt1p...` on regtest).
4. **Consensus Validation**: Verifying all derived addresses directly against Bitcoin Core using `validateaddress`.

---

## How to Run

Assumes a running regtest node (`bash scripts/init-lab.sh`). One-time setup from
the repo root: Python — `python -m venv --copies .venv && .venv/bin/pip install -r
examples/python/requirements.txt`; TypeScript — `cd examples && npm install`;
Rust/Go fetch their pinned deps on first build.

### Python
```bash
.venv/bin/python examples/02-keys-and-addresses/python/main.py
```

### TypeScript
```bash
cd examples && npx tsx 02-keys-and-addresses/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/02-keys-and-addresses/rust/Cargo.toml
```

### Go
```bash
cd examples/02-keys-and-addresses/go && go run .
```

---

## Expected Output

Derived from the canonical `abandon abandon … about` mnemonic on regtest; every
language port asserts these exact values (and cross-checks the fingerprint
`73c5da0a`).

```text
=== Lab 02: Keys & Addresses ===
[Step 1] Bootstrapping lab wallet & RPC connection ... ✓
[Step 2] Generating BIP39 mnemonic & root seed ... ✓ (12 words, 64-byte seed)
[Step 3] Deriving BIP32 root key ... ✓ (fingerprint: 73c5da0a)
[Step 4] Deriving BIP44 Legacy P2PKH address ... ✓ (mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV)
[Step 5] Deriving BIP84 SegWit P2WPKH address ... ✓ (bcrt1q6rz28mcfaxtmd6v789l9rrlrusdprr9pz3cppk)
[Step 6] Deriving BIP86 Taproot P2TR address ... ✓ (bcrt1p8wpt9v4frpf3tkn0srd97pksgsxc5hs52lafxwru9kgeephvs7rqjeprhg)
[Step 7] Validating addresses against Bitcoin Core node ... ✓ (all addresses valid)
======================================================
Result: PASS ✓
======================================================
```
