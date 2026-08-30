# 00 · CLI Workshop (No Code)

A 10-minute, no-code crash course: from a fresh regtest node to a confirmed
transaction using only `bitcoin-cli` (aliased as `btc` in the sandbox).

Start here if you are new to the command line or just want to see the node
working before diving into the language examples.

**Prerequisites:** the lab node is running. In a fresh clone:

```bash
bash scripts/init-lab.sh
```

`btc` is `bitcoin-cli -datadir=$HOME/btc-lab`. Outside the devcontainer, substitute
your own `bitcoin-cli` invocation.

---

## Step 1: Health Check

```bash
btc getblockchaininfo
```

A fresh node reports `"chain": "regtest"` and `"blocks": 0`.

## Step 2: Create a Test Wallet

```bash
btc createwallet "lab"
```

## Step 3: Generate a Receiving Address

```bash
ADDR=$(btc getnewaddress)
echo $ADDR
```

## Step 4: Mine Blocks to Mature the Subsidy (Coinbase Maturity)

```bash
btc generatetoaddress 110 "$ADDR"
```

> **Why 110 blocks?** Mining rewards cannot be spent until they are 100 blocks
> deep. Mining 110 blocks matures the first 10 block rewards
> (10 × 50 BTC = 500 BTC).

## Step 5: Inspect Balance vs. Discrete Coins (UTXOs)

```bash
btc getbalance
btc listunspent
```

> **The UTXO model:** Bitcoin has no "balance" table in the ledger. It tracks
> discrete Unspent Transaction Outputs (UTXOs). Your balance is simply the sum
> of all UTXOs your keys control.

## Step 6: Build, Sign & Broadcast a Transaction

```bash
DEST=$(btc getnewaddress)
TXID=$(btc sendtoaddress "$DEST" 1.0)
echo "Transaction ID: $TXID"
```

## Step 7: Inspect the Raw Transaction & Mempool

```bash
# View inputs (vin), recipient output, change output, and implicit fee
btc getrawtransaction "$TXID" true

# Inspect the waiting room for unconfirmed transactions
btc getrawmempool true
```

## Step 8: Confirm into a Block & Bury Deep

```bash
MINER=$(btc getnewaddress)
btc generatetoaddress 1 "$MINER"

# Confirmations = 1
btc gettransaction "$TXID" | grep confirmations

# Mine 5 more blocks to reach standard settlement depth (6 confirmations)
btc generatetoaddress 5 "$MINER"
btc gettransaction "$TXID" | grep confirmations
```

## Reset the Lab Cleanly Anytime

Regtest is 100% disposable. Reset back to block height 0 whenever you want:

```bash
./scripts/reset-lab.sh
```

---

## Experiments to Try

- **Spend before maturity:** mine only 50 blocks instead of 110, then try
  `sendtoaddress` and observe the error.
- **Watch change grow:** send several payments and inspect `btc listunspent`.
- **Fee adjustments:** change `fallbackfee` in `~/btc-lab/bitcoin.conf` and
  observe mempool fees.

---

## Command Cheatsheet

| Action | Command | Description |
| :--- | :--- | :--- |
| **Node info** | `btc getblockchaininfo` | Chain name, block count, headers |
| **Mining** | `btc generatetoaddress <n> <addr>` | Mine `<n>` blocks immediately |
| **Wallet** | `btc createwallet <name>` | Create a new descriptor wallet |
| **Address** | `btc getnewaddress` | Derive a fresh receiving address |
| **Balance** | `btc getbalance` | Total spendable balance |
| **UTXOs** | `btc listunspent` | List unspent outputs with amounts & confirmations |
| **Send** | `btc sendtoaddress <addr> <amount>` | Construct, sign, and broadcast a payment |
| **Inspect tx** | `btc getrawtransaction <txid> true` | Decode inputs, outputs, scripts, and fees |
| **Mempool** | `btc getrawmempool true` | View pending unconfirmed transactions |
| **Reset** | `./scripts/reset-lab.sh` | Wipe chain state and restart the node cleanly |
