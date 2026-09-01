# 02 · Keys & Addresses (Cryptographic Primitives & Derivation)

This lab explores cryptographic key derivation in Bitcoin — from a BIP39 mnemonic seed phrase to BIP32/BIP44 Hierarchical Deterministic (HD) wallets and modern address encodings (Legacy P2PKH, Native SegWit BIP84 P2WPKH, and Taproot BIP86 P2TR).

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**:
  - [Chapter 4: Keys and Addresses](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch04_keys.adoc)
  - [Chapter 5: Wallets](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch05_wallets.adoc)
- **Mastering the Lightning Network**:
  - [Chapter 4: Lightning Network Architecture](https://github.com/lnbook/lnbook/blob/develop/ch04_node_architecture.adoc)

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

### Python
```bash
python examples/02-keys-and-addresses/python/main.py
```

### TypeScript
```bash
npx tsx examples/02-keys-and-addresses/typescript/main.ts
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
```text
=== Lab 02: Keys & Addresses ===
[Step 1] Bootstrap lab wallet & RPC connection ... ✓
[Step 2] Generate BIP39 mnemonic & binary seed ... ✓ (12 words)
[Step 3] Derive BIP32 master root key ... ✓
[Step 4] Derive BIP44 Legacy P2PKH address ... ✓ (addr: m...)
[Step 5] Derive BIP84 Native SegWit P2WPKH address ... ✓ (addr: bcrt1q...)
[Step 6] Derive BIP86 Taproot P2TR address ... ✓ (addr: bcrt1p...)
[Step 7] Validate derived addresses with Bitcoin Core node ... ✓
======================================================
Result: PASS ✓
======================================================
```
