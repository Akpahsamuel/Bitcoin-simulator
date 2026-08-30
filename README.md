# Intro to Bitcoin · Part 2 — Workshop (Virtual Lab)

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/Akpahsamuel/Bitcoin-simulator?quickstart=1)
[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/Akpahsamuel/Bitcoin-simulator)

> **"Run a node. Build a transaction."**  
> A stepwise, practical workshop for beginners in Bitcoin development using **Bitcoin Core 31.1** in `regtest` mode.

This repository provides a zero-install cloud environment accompanying the slide deck **`IntroToBitcoin-Part2-Workshop.pptx`**. You can run every command in your browser via **GitHub Codespaces** or **Gitpod** without downloading the Bitcoin blockchain or compiling software.

---

## Quick Start (Cloud Lab)

1. Launch your lab using the **[Open in GitHub Codespaces](https://codespaces.new/Akpahsamuel/Bitcoin-simulator?quickstart=1)** or **[Open in Gitpod](https://gitpod.io/#https://github.com/Akpahsamuel/Bitcoin-simulator)** badge.
2. In ~45–60 seconds, your private container opens with `bitcoind` already running in `regtest` mode.
3. Test your node:
   ```bash
   btc getblockchaininfo
   ```
4. Follow the workshop steps below!

---

## Workshop Walkthrough (Slides 1–16)

The lab provides a preconfigured CLI helper:
```bash
# Pre-installed helper pointing to your node datadir:
btc <command>

# Equivalent Slide 5 alias:
alias btc='bitcoin-cli -datadir=$HOME/btc-lab'
```

---

### Step 1: Configure Your Node (Slide 4)
*Already pre-configured in your lab container at `~/btc-lab/bitcoin.conf`.*

```ini
regtest=1
[regtest]
fallbackfee=0.0002
txindex=1
```

> **Concept:** A tiny configuration file tells Bitcoin Core to run in `regtest` (a local, private sandbox with instant mining and no real money).

---

### Step 2: Start the Node (Slide 5)
*Already started automatically on lab launch.*

```bash
bitcoind -datadir=$HOME/btc-lab -daemon
btc getblockchaininfo
```

Expected output:
```json
{
  "chain": "regtest",
  "blocks": 0
}
```

---

### Step 3: Create a Wallet (Slide 6)
Make a wallet to hold your cryptographic keys.

```bash
btc createwallet "lab"
```

> **Concept:** A Bitcoin wallet is a keychain holding private keys that authorize spending.

---

### Step 4: Get a Receiving Address (Slide 7)
Ask the wallet for an address to receive coins.

```bash
btc getnewaddress
```

> **Concept:** The wallet derives a key pair, hashes the public key, and encodes it as a `bcrt1...` address. An address is safe to share publicly.

---

### Step 5: Mine Some Coins (Slide 8)
On regtest, you create blocks instantly on demand.

```bash
ADDR=$(btc getnewaddress)
btc generatetoaddress 110 "$ADDR"
```

> **Concept (Coinbase Maturity):** Each block rewards the miner a 50 BTC subsidy. Block rewards cannot be spent until they are **100 blocks deep** (maturity rule). Mining 110 blocks matures the first 10 rewards.

---

### Step 6: Check Your Balance (Slide 9)
See how much spendable bitcoin you now have.

```bash
btc getbalance
```

Output:
```
500.00000000
```

> **Concept:** 10 matured rewards × 50 BTC = 500 BTC. Blocks 11–110 are younger than 100 blocks, so their rewards are not yet spendable.

---

### Step 7: You Don't Have a Balance — You Have Coins (Slide 10)
List the actual unspent outputs your wallet controls.

```bash
btc listunspent
```

> **Concept (The UTXO Model):** Bitcoin has no "balance" field in the blockchain state. It only tracks **Unspent Transaction Outputs (UTXOs)** — discrete coins like physical cash notes. Your balance is simply the sum of your UTXOs (10 × 50 BTC = 500 BTC).

---

### Step 8: Send Your First Transaction (Slide 11)
Pay 1 BTC to a new destination address.

```bash
DEST=$(btc getnewaddress)
TXID=$(btc sendtoaddress "$DEST" 1.0)
echo $TXID
```

> **Concept:** `sendtoaddress` performs three jobs in one command:
> 1. Selects which UTXO(s) to spend
> 2. Signs the transaction with your private key
> 3. Broadcasts it to the local mempool, returning the unique `txid`

---

### Step 9: Look Inside the Transaction (Slide 12)
Decode the raw transaction to see what actually happened.

```bash
btc getrawtransaction "$TXID" true
```

Look at the structure:
- **`vin` (Inputs):** The one 50 BTC coin spent.
- **`vout[0]` (Recipient Output):** 1.00000000 BTC to `$DEST`.
- **`vout[1]` (Change Output):** ≈ 48.99997 BTC back to your wallet.
- **Fee:** Inputs minus outputs (≈ 0.00003 BTC) paid to the miner.

---

### Step 10: Where Does It Wait? The Mempool (Slide 13)
Inspect your transaction waiting in the memory pool.

```bash
btc getrawmempool true
```

> **Concept:** The mempool is a node's waiting room for valid, unconfirmed transactions awaiting inclusion in a block.

---

### Step 11: Confirm Your Transaction (Slide 14)
Mine a block to include the transaction, then bury it deeper under additional blocks.

```bash
MINER=$(btc getnewaddress)
btc generatetoaddress 1 "$MINER"

# Check confirmations (now 1)
btc gettransaction "$TXID" | grep confirmations

# Mine 5 more blocks to bury it deeper
btc generatetoaddress 5 "$MINER"
btc gettransaction "$TXID" | grep confirmations
```

> **Concept:** Mining a block containing your transaction gives it **1 confirmation**. Each subsequent block stacked on top adds another confirmation. Deeper blocks require exponentially more work to reverse.

---

## Reset & Experiment Again (Slide 16)

Regtest is completely disposable. Wipe the blockchain state anytime to run the pipeline again:

```bash
./scripts/reset-lab.sh
```

Or run the manual steps:
```bash
btc stop
rm -rf ~/btc-lab/regtest
bitcoind -datadir=$HOME/btc-lab -daemon
```

### Experiments to Try:
- **Spend before maturity:** Mine only 50 blocks instead of 110, then try `sendtoaddress` and observe the error.
- **Watch change grow:** Send several payments and inspect `btc listunspent`.
- **Fee adjustments:** Change `fallbackfee` in `bitcoin.conf` and observe mempool fees.

---

## Command Cheatsheet (Slide 18)

| Action | Command | Purpose |
| :--- | :--- | :--- |
| **Health Check** | `btc getblockchaininfo` | Verify chain & block height |
| **Wallet** | `btc createwallet "lab"` | Initialize a keychain |
| **Address** | `btc getnewaddress` | Generate a new receiving address |
| **Mine / Fund** | `btc generatetoaddress 110 "$ADDR"` | Mine blocks to mature coinbase rewards |
| **Balance** | `btc getbalance` | Check spendable total |
| **Coins (UTXOs)** | `btc listunspent` | View individual unspent outputs |
| **Send** | `btc sendtoaddress "$DEST" 1.0` | Build, sign & broadcast payment |
| **Inspect** | `btc getrawtransaction "$TXID" true` | View inputs, outputs, and change |
| **Mempool** | `btc getrawmempool true` | Inspect unconfirmed transactions |
| **Confirm** | `btc generatetoaddress 1 "$MINER"` | Mine transaction into a block |
| **Reset** | `./scripts/reset-lab.sh` | Wipe chain and restart clean |

---

## Contributing & Branch Protection

We welcome contributions! Please note that the repository enforces GitHub repository protection rules on `main`:
- **Pull Requests Only**: Direct pushes to `main` are blocked.
- **Verified Commits Required**: All commits must be cryptographically signed (GPG or SSH). Unverified commits will block merge checks.

Check out [**CONTRIBUTING.md**](CONTRIBUTING.md) for quick step-by-step setup guides on configuring SSH/GPG commit signing.
