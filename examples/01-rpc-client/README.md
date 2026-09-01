# 01 · RPC Client (Foundational Node Interfacing)

This lab demonstrates how client backends authenticate with Bitcoin Core, query chain state, manage wallets, trigger mining, and access unauthenticated REST endpoints.

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**: [Chapter 3: Bitcoin Core: The Reference Implementation](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch03_bitcoin-core.adoc)
- **Mastering the Lightning Network**: [Chapter 1: Introduction](https://github.com/lnbook/lnbook/blob/develop/ch01_intro.adoc)

---

## Concepts Demonstrated
1. **JSON-RPC Authentication**: Connecting to Bitcoin Core over HTTP Basic Authentication (`bitcoinrpc:bitcoinrpcpassword`).
2. **Blockchain & Node Inspection**: Calling `getblockchaininfo` to inspect verification progress, block height, and chain parameters.
3. **Wallet Creation & Address Generation**: Calling `createwallet` / `loadwallet` and `getnewaddress`.
4. **On-Demand Mining**: Driving regtest block progression using `generatetoaddress`.
5. **High-Performance REST Interface**: Querying Bitcoin Core's unauthenticated REST endpoint `/rest/chaininfo.json`.

---

## How to Run

### Python
```bash
python examples/01-rpc-client/python/main.py
```

### TypeScript
```bash
npx tsx examples/01-rpc-client/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/01-rpc-client/rust/Cargo.toml
```

### Go
```bash
cd examples/01-rpc-client/go && go run .
```

---

## Expected Output
```text
=== Lab 01: RPC Client ===
[Step 1] Bootstrap lab wallet & initial funds ... ✓
[Step 2] Query getblockchaininfo via JSON-RPC ... ✓ (chain: regtest, blocks: >= 101)
[Step 3] Generate fresh address & mine 1 block ... ✓ (new block height: >= 102)
[Step 4] Query wallet balance via getbalance ... ✓ (balance > 0 BTC)
[Step 5] Query unauthenticated REST API (/rest/chaininfo.json) ... ✓
======================================================
Result: PASS ✓
======================================================
```
