# Bitcoin Developer Sandbox & Starter Kit

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/Akpahsamuel/Bitcoin-simulator?quickstart=1)
[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/Akpahsamuel/Bitcoin-simulator)
![Bitcoin Core](https://img.shields.io/badge/Bitcoin%20Core-31.1-orange?logo=bitcoin)
![Network](https://img.shields.io/badge/Network-Regtest-blue)
![Languages](https://img.shields.io/badge/Languages-Python%20%7C%20TypeScript%20%7C%20Rust%20%7C%20Go-brightgreen)
![License](https://img.shields.io/badge/License-MIT-lightgrey)

> **The ultimate interactive sandbox for learning Bitcoin development, building on-chain and Layer-2 applications, and preparing to contribute to Bitcoin Core.**

This repository turns concepts from **Mastering Bitcoin (3rd Edition)** and **Mastering the Lightning Network** into an instant, zero-install, hands-on playground. It ships a pre-configured **Bitcoin Core 31.1 regtest node** with **JSON-RPC, REST, and ZeroMQ (ZMQ)** enabled, a language-agnostic example suite (**Python & TypeScript** as full references, **Rust & Go** starter ports), and chapter-by-chapter links into both companion books.

---

## Table of Contents

- [Why This Sandbox?](#why-this-sandbox)
- [Quick Start](#quick-start)
- [Sandbox Architecture & Endpoints](#sandbox-architecture--endpoints)
- [Study Guides & Book References](#study-guides--book-references)
- [Example Projects Roadmap (`examples/`)](#example-projects-roadmap)
- [Hands-On Workshop Lab (Fast-Track)](#hands-on-workshop-lab-fast-track)
- [Contributing to Bitcoin Core & Ecosystem](#contributing-to-bitcoin-core--ecosystem)
- [Contributing to This Repository](#contributing-to-this-repository)
- [Command Cheatsheet](#command-cheatsheet)

---

## Why This Sandbox?

Learning Bitcoin development often hits immediate friction: syncing a multi-gigabyte blockchain, configuring RPC credentials, handling wallet descriptors, and risking real coins.

This sandbox eliminates all setup hurdles:
- **Instant Cloud Environment**: Spin up a complete Linux development container in ~45 seconds via GitHub Codespaces or Gitpod.
- **Zero Blockchain Download**: Runs on private **regtest** (regression testing mode) where you mine blocks instantly on demand with zero delay and zero cost.
- **Modern Bitcoin Toolchain**: Bitcoin Core 31.1 with descriptor wallets, Taproot (BIP340/341/342), SegWit (BIP84/BIP86), REST API, and ZeroMQ streaming pre-enabled.
- **Theory Mapped to Execution**: Every example is cross-referenced to specific chapters of the open-source editions of *Mastering Bitcoin* and *Mastering the Lightning Network* (linked below; optionally cloned locally).
- **Equipping Builders & Contributors**: Learn both how to build consumer-facing Bitcoin applications (wallets, payment processors, escrow contracts) and how Bitcoin Core operates internally under the hood.

---

## Quick Start

### 1. Launch in Browser (Recommended)
Click **[Open in GitHub Codespaces](https://codespaces.new/Akpahsamuel/Bitcoin-simulator?quickstart=1)** or **[Open in Gitpod](https://gitpod.io/#https://github.com/Akpahsamuel/Bitcoin-simulator)**.

Your terminal opens automatically with Bitcoin Core initialized and running in regtest.

### 2. Verify Node Health
```bash
btc getblockchaininfo
```

You should see:
```json
{
  "chain": "regtest",
  "blocks": 0,
  "verificationprogress": 1
}
```

### 3. Run Locally with Docker / VS Code Dev Containers
If you prefer developing locally on your machine:
1. Clone the repo and open it in **VS Code**.
2. Install the **Dev Containers** extension (`ms-vscode-remote.remote-containers`).
3. Press `Cmd+Shift+P` (macOS) or `F1` and select: **Dev Containers: Reopen in Container**.

---

## Sandbox Architecture & Endpoints

Your node runs inside the container with full developer services exposed:

| Service | Port | Protocol | Purpose / Description |
| :--- | :--- | :--- | :--- |
| **JSON-RPC** | `18443` | HTTP / JSON-RPC | Full node management, wallet manipulation, transaction broadcast |
| **REST API** | `18443` | HTTP / REST | Unauthenticated, fast endpoints (`/rest/chaininfo.json`, `/rest/block/`, `/rest/tx/`) |
| **ZMQ Blocks** | `28332` | TCP / ZeroMQ (`pubrawblock`) | Real-time block discovery publisher for backend event loops |
| **ZMQ Txs** | `28333` | TCP / ZeroMQ (`pubrawtx`) | Real-time mempool transaction streamer |
| **P2P Network** | `18444` | TCP / Bitcoin Wire | Node-to-node peer communication |

### Built-in RPC Credentials
External SDKs and scripts can connect to the local node using:
- **Host**: `127.0.0.1` (or `localhost`)
- **Port**: `18443`
- **Username**: `bitcoinrpc`
- **Password**: `bitcoinrpcpassword`

---

## Study Guides & Book References

This sandbox is a companion to two open-source books. They are **not vendored into this repo** — the links below point upstream. To read them offline, clone them next to this repo (these paths are already in `.gitignore`):

```bash
git clone https://github.com/bitcoinbook/bitcoinbook.git
git clone https://github.com/lnbook/lnbook.git
```

### 📖 Mastering Bitcoin (3rd Edition)
Andreas M. Antonopoulos & David A. Harding — [full repository](https://github.com/bitcoinbook/bitcoinbook). Key chapters for hands-on practice:
- [**Chapter 3: Bitcoin Core — The Reference Implementation**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch03_bitcoin-core.adoc) — Client commands, RPC interfaces, and configuration.
- [**Chapter 4: Keys & Addresses**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch04_keys.adoc) — Elliptic curves (secp256k1), WIF, Base58Check, and Bech32/Bech32m.
- [**Chapter 5: Wallets**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch05_wallets.adoc) — BIP32 HD derivation, BIP39 seed phrases, BIP43/44/84/86 paths.
- [**Chapter 6: Transactions**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch06_transactions.adoc) — Inputs, outputs, UTXO lifecycle, transaction fees, and serialization.
- [**Chapter 7: Authorization & Authentication**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch07_authorization-authentication.adoc) — Script execution, `OP_CHECKSIG`, SegWit witness, Taproot/Tapscript.
- [**Chapter 9: Fees**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch09_fees.adoc) — Fee rates (sat/vB), Replace-By-Fee (RBF, BIP125), Child-Pays-for-Parent (CPFP).
- [**Chapter 10: The Bitcoin Network**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch10_network.adoc) — P2P protocol, mempool propagation, block relay.
- [**Chapter 11: The Blockchain**](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch11_blockchain.adoc) — Block headers, Merkle trees, and chain reorganizations.

### ⚡ Mastering the Lightning Network
Andreas M. Antonopoulos, Olaoluwa Osuntokun & René Pickhardt — [full repository](https://github.com/lnbook/lnbook). Referenced for conceptual grounding only — **v1 ships no runnable Lightning labs**:
- [**Chapter 6: Lightning Architecture**](https://github.com/lnbook/lnbook/blob/develop/06_lightning_architecture.asciidoc) — Network design and components.
- [**Chapter 7: Payment Channels**](https://github.com/lnbook/lnbook/blob/develop/07_payment_channels.asciidoc) — Channel mechanics and commitment transactions.
- [**Chapter 8: Routing & HTLCs**](https://github.com/lnbook/lnbook/blob/develop/08_routing_htlcs.asciidoc) — Hash Time-Locked Contracts and multi-hop routing.

---

## Example Projects Roadmap

> 🚧 **Status:** the example suite is under construction. The [workshop lab](#hands-on-workshop-lab-fast-track) below is fully working today.

Each project targets one core protocol primitive. **Python** and **TypeScript** are the reference implementations (all six); **Rust** and **Go** ship `01-rpc-client` plus templates so the remaining ports are straightforward contributions.

```
examples/
├── 01-rpc-client/         # Query node status, mine blocks, manage balances via JSON-RPC + REST
├── 02-keys-and-addresses/ # BIP39 mnemonics, BIP32 HD derivation, Native SegWit (BIP84) & Taproot (BIP86)
├── 03-raw-transactions/   # Coin selection, raw tx serialization, witness construction, fee management
├── 04-multisig-escrow/    # 2-of-3 P2WSH multi-party coordination and cooperative/dispute spending
├── 05-timelocks/          # Absolute (OP_CHECKLOCKTIMEVERIFY) and Relative (OP_CHECKSEQUENCEVERIFY) contracts
└── 06-zmq-listener/       # Real-time event streaming backend for blocks & mempool transactions
```

Each example lives at `examples/<NN-name>/<language>/` and reads its connection settings from a shared `.env` (generated from `.env.example` by `init-lab.sh`).

### Running the Example Suites
Once an example is available, run it directly (e.g. `python main.py`, `npm start`, `cargo run`, `go run .`) or run the whole suite:
```bash
bash scripts/test-examples.sh
```

---

## Hands-On Workshop Lab (Fast-Track)

New to the command line? This 10-minute crash course takes you from a fresh node to a confirmed transaction. Every step works today.

### Step 1: Health Check
```bash
btc getblockchaininfo
```

### Step 2: Create a Test Wallet
```bash
btc createwallet "lab"
```

### Step 3: Generate a Receiving Address
```bash
ADDR=$(btc getnewaddress)
echo $ADDR
```

### Step 4: Mine Blocks to Mature Subsidy (Coinbase Maturity)
```bash
btc generatetoaddress 110 "$ADDR"
```
> *Why 110 blocks?* Mining rewards cannot be spent until they are 100 blocks deep. Mining 110 blocks matures the first 10 block rewards (10 × 50 BTC = 500 BTC).

### Step 5: Inspect Balance vs. Discrete Coins (UTXOs)
```bash
btc getbalance
btc listunspent
```
> *The UTXO Model:* Bitcoin has no "balance" table in the ledger. It tracks discrete Unspent Transaction Outputs (UTXOs). Your balance is simply the sum of all UTXOs your keys control.

### Step 6: Build, Sign & Broadcast a Transaction
```bash
DEST=$(btc getnewaddress)
TXID=$(btc sendtoaddress "$DEST" 1.0)
echo "Transaction ID: $TXID"
```

### Step 7: Inspect the Raw Transaction & Mempool
```bash
# View inputs (vin), recipient output, change output, and implicit fee
btc getrawtransaction "$TXID" true

# Inspect the waiting room for unconfirmed transactions
btc getrawmempool true
```

### Step 8: Confirm into a Block & Bury Deep
```bash
MINER=$(btc getnewaddress)
btc generatetoaddress 1 "$MINER"

# Confirmations = 1
btc gettransaction "$TXID" | grep confirmations

# Mine 5 more blocks to reach standard settlement depth (6 confirmations)
btc generatetoaddress 5 "$MINER"
btc gettransaction "$TXID" | grep confirmations
```

### Reset the Lab Cleanly Anytime
Regtest is 100% disposable. Reset back to block height 0 whenever you want:
```bash
./scripts/reset-lab.sh
```

### Experiments to Try:
- **Spend before maturity:** Mine only 50 blocks instead of 110, then try `sendtoaddress` and observe the error.
- **Watch change grow:** Send several payments and inspect `btc listunspent`.
- **Fee adjustments:** Change `fallbackfee` in `bitcoin.conf` and observe mempool fees.

---

## Contributing to Bitcoin Core & Ecosystem

One of the primary purposes of this sandbox is to prepare developers to contribute to open-source Bitcoin repositories:

1. **Bitcoin Core (`bitcoin/bitcoin`)**:
   - Understand the RPC test framework used in Core's `test/functional/`.
   - Test bug reproductions and PR reviews in regtest before running full test suites.
2. **Bitcoin Dev Tooling & Libraries**:
   - Build and test wallet libraries (`bitcoinjs-lib`, `bip-utils`, `bdk`, `rust-bitcoin`).
   - Implement custom signers, hardware wallet bridges, or indexers using the ZMQ and REST interfaces.
3. **Lightning & L2 Protocols**:
   - Use this node as the local L1 backbone for Core Lightning (`lightningd`), LND, or Eclair nodes.

---

## Contributing to This Repository

We enthusiastically welcome contributions from developers worldwide!

> [!IMPORTANT]
> **Repository Rules on `main`**:
> 1. **Pull Requests Only**: Direct commits to `main` are rejected.
> 2. **Verified Commits Required**: **Every commit in your PR must be cryptographically signed (GPG or SSH).** Unsigned commits will fail branch protection checks and block merging.

Please read [**CONTRIBUTING.md**](CONTRIBUTING.md) for full instructions on setting up SSH/GPG commit signing and submitting PRs.

---

## Command Cheatsheet

| Action | Command | Description |
| :--- | :--- | :--- |
| **Node Info** | `btc getblockchaininfo` | Chain name, block count, headers |
| **Mining** | `btc generatetoaddress <n> <addr>` | Mine `<n>` blocks immediately |
| **Wallet** | `btc createwallet <name>` | Create a new descriptor wallet |
| **Address** | `btc getnewaddress` | Derive a fresh receiving address |
| **Balance** | `btc getbalance` | Total spendable balance |
| **UTXOs** | `btc listunspent` | List unspent outputs with amounts & confirmations |
| **Send** | `btc sendtoaddress <addr> <amount>` | Construct, sign, and broadcast payment |
| **Inspect Tx** | `btc getrawtransaction <txid> true` | Decode inputs, outputs, scripts, and fees |
| **Mempool** | `btc getrawmempool true` | View pending unconfirmed transactions |
| **Reset** | `./scripts/reset-lab.sh` | Wipe chain state and restart node cleanly |
