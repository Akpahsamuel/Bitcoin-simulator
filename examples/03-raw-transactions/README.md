# 03 · Raw Transactions (Manual UTXO Selection, Construction & Broadcast)

This lab takes you under the hood of Bitcoin transaction construction. Rather than relying on automatic wallet abstractions, you manually select UTXOs, calculate fees, build raw transaction byte structures, sign inputs, verify mempool acceptance rules, and broadcast to the network.

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**: [Chapter 6: Transactions](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch06_transactions.adoc)
- **Mastering the Lightning Network**: [Chapter 8: Routing & HTLCs](https://github.com/lnbook/lnbook/blob/develop/08_routing_htlcs.asciidoc)

---

## Concepts Demonstrated
1. **UTXO Model & Selection**: Querying `listunspent` to select specific transaction outputs as inputs.
2. **Fee Calculation & Change Allocation**: Computing sat/vB fees and returning leftover funds via a change output.
3. **Raw Transaction Construction**: Creating the serialized transaction payload with `createrawtransaction`.
4. **Cryptographic Signing**: Signing input witnesses with the wallet private keys via `signrawtransactionwithwallet`.
5. **Mempool Policy Pre-flight (`testmempoolaccept`)**: Testing acceptance against consensus and policy rules before network broadcast.
6. **Network Broadcast & Confirmation**: Submitting via `sendrawtransaction` and mining a block to verify confirmation depth.

---

## How to Run

Assumes a running regtest node (`bash scripts/init-lab.sh`). One-time setup from
the repo root: Python — `python -m venv --copies .venv && .venv/bin/pip install -r
examples/python/requirements.txt`; TypeScript — `cd examples && npm install`;
Rust/Go fetch their pinned deps on first build.

### Python
```bash
.venv/bin/python examples/03-raw-transactions/python/main.py
```

### TypeScript
```bash
cd examples && npx tsx 03-raw-transactions/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/03-raw-transactions/rust/Cargo.toml
```

### Go
```bash
cd examples/03-raw-transactions/go && go run .
```

---

## Expected Output
```text
=== Lab 03: Raw Transactions ===
[Step 1] Bootstrap lab wallet & ensure spendable UTXOs ... ✓
[Step 2] Select UTXO from listunspent ... ✓ (txid: ..., vout: ..., amount: 50.0 BTC)
[Step 3] Construct raw transaction hex (createrawtransaction) ... ✓
[Step 4] Sign transaction inputs (signrawtransactionwithwallet) ... ✓ (complete: true)
[Step 5] Verify transaction with testmempoolaccept ... ✓ (allowed: true)
[Step 6] Broadcast transaction via sendrawtransaction ... ✓ (txid: ...)
[Step 7] Mine 1 block & verify confirmation ... ✓ (confirmations: 1)
======================================================
Result: PASS ✓
======================================================
```
