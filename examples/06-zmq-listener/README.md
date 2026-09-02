# 06 · ZMQ Listener (Event-Driven Backends)

Bitcoin Core pushes real-time notifications over **ZeroMQ**: every mempool
transaction on `zmqpubrawtx`, every connected block on `zmqpubrawblock`. This is
how an indexer, a payment processor, or a Lightning node reacts to chain
activity without polling `getblockcount` in a loop.

This lab subscribes to both streams, triggers one transaction and one block,
receives the raw bytes, and verifies them against the node:

```
rawtx    → decoderawtransaction(payload).txid  == the txid we broadcast
rawblock → getblock(minedHash, 0)              == payload (hex)
```

If nothing arrives within the timeout the lab **fails loudly** — that is the
"silent ZMQ" symptom (wrong port, `zmqpub*` not set, or bound to the wrong
interface); see [`docs/troubleshooting.md`](../../docs/troubleshooting.md).

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**: [Chapter 10: The Bitcoin Network](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch10_network.adoc) — transaction and block propagation.
- **Mastering the Lightning Network**: [Chapter 5: Node Operations](https://github.com/lnbook/lnbook/blob/develop/05_node_operations.asciidoc) — nodes watch the chain via exactly this kind of event feed.

---

## Concepts Demonstrated
1. **ZMQ SUB sockets**: connecting to `tcp://…:28332` / `:28333` (from `.env`), subscribing to the `rawblock` / `rawtx` topics.
2. **Multipart frames**: each notification is `[topic, payload, sequence]`; the 4-byte little-endian sequence lets a consumer detect dropped messages.
3. **Receive timeouts**: a bounded wait so a misconfigured feed is a test failure, not a hang.
4. **Round-trip verification**: the pushed bytes are validated back through the JSON-RPC interface.

---

## How to Run

Assumes a running regtest node (`bash scripts/init-lab.sh`) and the one-time
setup (`bash scripts/setup-examples.sh`).

### Python
```bash
.venv/bin/python examples/06-zmq-listener/python/main.py
```

### TypeScript
```bash
cd examples && npx tsx 06-zmq-listener/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/06-zmq-listener/rust/Cargo.toml
```

### Go
```bash
cd examples/06-zmq-listener/go && go run .
```

---

## Expected Output
```text
=== Lab 06: ZMQ Listener ===
[Step 1] Bootstrapping lab wallet & funds ... ✓
[Step 2] Subscribing to rawtx (…:28333) and rawblock (…:28332) ... ✓
[Step 3] Broadcasting a transaction ... ✓ (txid: …)
[Step 4] Received rawtx frame & verified txid ... ✓
[Step 5] Mining 1 block ... ✓ (hash: …)
[Step 6] Received rawblock frame & verified against getblock ... ✓
======================================================
Result: PASS ✓
======================================================
```
