# Bitcoin Workshop: Virtual Lab (Bitcoin Core 31.1)

<a href="https://codespaces.new/Akpahsamuel/Bitcoin-simulator?quickstart=1" target="_blank" rel="noopener noreferrer"><img src="https://github.com/codespaces/badge.svg" alt="Open in GitHub Codespaces"></a>
<a href="https://gitpod.io/#https://github.com/Akpahsamuel/Bitcoin-simulator" target="_blank" rel="noopener noreferrer"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod"></a>

This repository provides a zero-install, pre-configured cloud environment for running the **Intro to Bitcoin · Part 2 — Workshop** (`IntroToBitcoin-Part2-Workshop.pptx`).

If you cannot or do not wish to install Bitcoin Core locally on your laptop, launch this virtual lab directly in your browser.

---

## Quick Start (No Local Install Needed)

1. Click the **Open in GitHub Codespaces** badge above (or launch via Gitpod).
2. Wait ~45–60 seconds for your private container to build and start.
3. The integrated terminal will open with **Bitcoin Core 31.1** already running in `regtest` mode.
4. Verify your node is ready:
   ```bash
   btc getblockchaininfo
   ```
   You should see:
   ```json
   {
     "chain": "regtest",
     "blocks": 0
   }
   ```
5. Follow the workshop slides from **Slide 6** onwards, or run through Slides 4–5 if you want to inspect configuration.

---

## How It Works

- **Software**: Official Bitcoin Core 31.1 binaries (`bitcoind` and `bitcoin-cli`) running on `debian:bookworm-slim`.
- **Environment**: A private, disposable `regtest` (regression test) network running locally in your container.
- **CLI Alias**: The `btc` helper is pre-installed on your `PATH` and points to your node's datadir (`$HOME/btc-lab`). You can also run Slide 5's alias verbatim:
  ```bash
  alias btc='bitcoin-cli -datadir=$HOME/btc-lab'
  ```
- **Idempotent Setup**: Slide 4 (`bitcoin.conf`) and Slide 5 (`bitcoind -daemon`) are already pre-initialized, but you can safely run those commands without breaking anything.

---

## Workshop Step-by-Step Cheatsheet

All commands match the workshop deck byte-for-byte:

| Step | Slide | Purpose | Command |
| :--- | :--- | :--- | :--- |
| **Health Check** | Slide 5 | Check chain & block height | `btc getblockchaininfo` |
| **Wallet** | Slide 6 | Create keychain | `btc createwallet "lab"` |
| **Address** | Slide 7 | Generate receiving address | `btc getnewaddress` |
| **Mine / Fund** | Slide 8 | Mine 110 blocks to mature subsidy | `ADDR=$(btc getnewaddress)`<br>`btc generatetoaddress 110 "$ADDR"` |
| **Balance** | Slide 9 | Check spendable coins (500 BTC) | `btc getbalance` |
| **UTXOs** | Slide 10 | List unspent transaction outputs | `btc listunspent` |
| **Send** | Slide 11 | Build, sign & broadcast payment | `DEST=$(btc getnewaddress)`<br>`TXID=$(btc sendtoaddress "$DEST" 1.0)`<br>`echo $TXID` |
| **Inspect** | Slide 12 | Inspect raw transaction details | `btc getrawtransaction "$TXID" true` |
| **Mempool** | Slide 13 | View pending mempool transactions | `btc getrawmempool true` |
| **Confirm** | Slide 14 | Mine a block to confirm the tx | `MINER=$(btc getnewaddress)`<br>`btc generatetoaddress 1 "$MINER"` |

---

## Resetting the Lab

Regtest is completely disposable. Whenever you want to wipe the blockchain and start fresh (as shown on **Slide 16**), run:

```bash
./scripts/reset-lab.sh
```

Or run the manual reset commands from Slide 16:
```bash
btc stop
rm -rf ~/btc-lab/regtest
bitcoind -datadir=$HOME/btc-lab -daemon
```

---

## Running Locally via VS Code Dev Containers

If you have **Docker Desktop** and **VS Code** with the *Dev Containers* extension installed on your laptop:
1. Clone this repository or open the project folder in VS Code.
2. Press `F1` (or `Cmd+Shift+P` on macOS) and select:
   **Dev Containers: Reopen in Container**.
3. VS Code will build the Docker container and open the terminal ready to use.

