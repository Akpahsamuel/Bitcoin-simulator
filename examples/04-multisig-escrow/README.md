# 04 · Multisig Escrow (2-of-3 P2WSH, Manual Witness Assembly)

This lab builds a **2-of-3 native SegWit multisig** from scratch — no
`addmultisigaddress`, no wallet multisig helpers (Bitcoin Core 31 is
descriptor-wallet only and those RPCs are gone). Three parties each hold a key;
any two can move the funds. You construct the witness script, derive the P2WSH
address, fund it, then **assemble the witness stack by hand** to spend it:

```
witness = [ <empty>, <sig from key A>, <sig from key B>, <witnessScript> ]
```

The leading empty item is the well-known `OP_CHECKMULTISIG` off-by-one dummy.

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**: [Chapter 7: Authorization & Authentication](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch07_authorization-authentication.adoc) — `OP_CHECKMULTISIG`, P2SH/P2WSH, witness construction.
- **Mastering the Lightning Network**: [Chapter 7: Payment Channels](https://github.com/lnbook/lnbook/blob/develop/07_payment_channels.asciidoc) — the 2-of-2 funding output is the same P2WSH machinery.

---

## Concepts Demonstrated
1. **Raw multisig script**: `OP_2 <pubA> <pubB> <pubC> OP_3 OP_CHECKMULTISIG` built with the library's script builder, not an RPC.
2. **P2WSH address**: witness v0, `OP_0 <sha256(witnessScript)>`, Bech32 (`bcrt1q…`, 32-byte program).
3. **Funding** a script address from the `lab` descriptor wallet and locating the resulting UTXO by its `scriptPubKey`.
4. **BIP143 sighash**: the SegWit v0 signature-hash algorithm, with the *witnessScript* as scriptCode.
5. **Threshold signing**: two of the three keys produce DER signatures with a `SIGHASH_ALL` byte appended.
6. **Manual witness assembly**: placing `["", sigA, sigB, witnessScript]` on the input, serialising, and broadcasting via `testmempoolaccept` + `sendrawtransaction`.

All four ports derive the **same** script and address for the fixed test keys
(`0x11…11`, `0x22…22`, `0x33…33`) and assert it:

```
witnessScript  5221034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa
               2102466d7fcae563e5cb09a0d1870bb580344804617879a14949cf22285f1bae3f27
               21023c72addb4fdf09af94f0c94d7fe92a386a7e70cf8a1d85916386bb2535c7b1b153ae
P2WSH address  bcrt1qpy8yjjs2l5neewx722mxve9w6m77zqsu7rldukggseflhwralerqh6ma0d
```

---

## How to Run

Assumes a running regtest node (`bash scripts/init-lab.sh`) and the one-time
setup (`bash scripts/setup-examples.sh`).

### Python
```bash
.venv/bin/python examples/04-multisig-escrow/python/main.py
```

### TypeScript
```bash
cd examples && npx tsx 04-multisig-escrow/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/04-multisig-escrow/rust/Cargo.toml
```

### Go
```bash
cd examples/04-multisig-escrow/go && go run .
```

---

## Expected Output
```text
=== Lab 04: Multisig Escrow ===
[Step 1] Bootstrapping lab wallet & funds ... ✓
[Step 2] Deriving 3 escrow keypairs from test vectors ... ✓
[Step 3] Building 2-of-3 witness script ... ✓ (matches canonical)
[Step 4] Deriving P2WSH address & validating with node ... ✓ (bcrt1qpy8y…)
[Step 5] Funding the multisig (0.5 BTC) & mining 1 block ... ✓ (txid: …)
[Step 6] Locating the funding UTXO by scriptPubKey ... ✓ (vout: …, 0.5 BTC)
[Step 7] Building the spend transaction ... ✓
[Step 8] BIP143 sighash + signing with keys A and B ... ✓
[Step 9] Assembling witness [<empty>, sigA, sigB, script] ... ✓
[Step 10] testmempoolaccept ... ✓ (allowed: true)
[Step 11] Broadcasting & mining 1 block ... ✓ (confirmations: 1)
======================================================
Result: PASS ✓
======================================================
```
