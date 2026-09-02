# 05 · Timelocks (CLTV Absolute + CSV Relative)

This lab builds two timelocked P2WSH outputs and proves, on a live node, that
each one is **unspendable early and spendable once its lock matures**:

| | Opcode | BIP | Lock is on | Enforced by |
| :-- | :-- | :-- | :-- | :-- |
| **Absolute** | `OP_CHECKLOCKTIMEVERIFY` (`OP_CLTV`) | 65 | the transaction's `nLockTime` | a wall-clock block height / time |
| **Relative** | `OP_CHECKSEQUENCEVERIFY` (`OP_CSV`) | 112 / 68 | the spending input's `nSequence` | confirmations since the UTXO was mined |

For each: build the script → derive the P2WSH address → fund it → **try to spend
too early and watch `testmempoolaccept` reject it** → advance the chain → spend
successfully.

---

## Educational Companion
- **Mastering Bitcoin (3rd Edition)**: [Chapter 7: Authorization & Authentication](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch07_authorization-authentication.adoc) — `OP_CLTV`, `OP_CSV`, `nLockTime`, `nSequence`, BIP68/65/112.
- **Mastering the Lightning Network**: [Chapter 7: Payment Channels](https://github.com/lnbook/lnbook/blob/develop/07_payment_channels.asciidoc) — relative timelocks are how a channel's revocation window works.

---

## Concepts Demonstrated
1. **CLTV script**: `<height> OP_CHECKLOCKTIMEVERIFY OP_DROP <pubkey> OP_CHECKSIG`.
2. **CSV script**: `<n> OP_CHECKSEQUENCEVERIFY OP_DROP <pubkey> OP_CHECKSIG` (relative block delay `n`).
3. **Transaction finality**: setting `nLockTime` + a non-final `nSequence` for CLTV; a version-2 tx with a BIP68 `nSequence` for CSV.
4. **Negative testing**: `testmempoolaccept` returns `allowed: false` (`non-final` / `non-BIP68-final`) for the early spend — the same transaction that later succeeds.
5. **BIP143 sighash + witness**: single-key `[<sig>, <witnessScript>]` stack, assembled by hand.

The CSV script is fixed (relative delay = 3 blocks) and every port asserts it:

```
CSV script   53b27521034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aaac
CSV address  bcrt1ql739fkda7sf20qkdwgku2j0ppeff4r7vsqasvvxestsqwtvuak3s9rktmg
```

The CLTV lock height is chosen at runtime (`current height + 10`) so the lab
stays re-runnable; the node still validates the derived address.

---

## How to Run

Assumes a running regtest node (`bash scripts/init-lab.sh`) and the one-time
setup (`bash scripts/setup-examples.sh`).

### Python
```bash
.venv/bin/python examples/05-timelocks/python/main.py
```

### TypeScript
```bash
cd examples && npx tsx 05-timelocks/typescript/main.ts
```

### Rust
```bash
cargo run --manifest-path examples/05-timelocks/rust/Cargo.toml
```

### Go
```bash
cd examples/05-timelocks/go && go run .
```

---

## Expected Output
```text
=== Lab 05: Timelocks ===
[Step 1] Bootstrapping lab wallet & funds ... ✓
[Step 2] CLTV: building <height> OP_CLTV script (lock at height N) ... ✓
[Step 3] CLTV: deriving P2WSH address & validating with node ... ✓
[Step 4] CLTV: funding (0.2 BTC) & mining 1 block ... ✓
[Step 5] CLTV: early spend rejected by testmempoolaccept ... ✓ (allowed: false — non-final)
[Step 6] CLTV: mining to the lock height ... ✓
[Step 7] CLTV: spend accepted, broadcast & confirmed ... ✓ (confirmations: 1)
[Step 8] CSV: building <3> OP_CSV script ... ✓ (matches canonical)
[Step 9] CSV: deriving P2WSH address & validating with node ... ✓
[Step 10] CSV: funding (0.2 BTC) & mining 1 block ... ✓
[Step 11] CSV: early spend rejected by testmempoolaccept ... ✓ (allowed: false — non-BIP68-final)
[Step 12] CSV: mining 3 blocks to satisfy the relative delay ... ✓
[Step 13] CSV: spend accepted, broadcast & confirmed ... ✓ (confirmations: 1)
======================================================
Result: PASS ✓
======================================================
```
