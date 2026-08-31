# Troubleshooting Guide

Common issues encountered when developing against local Bitcoin Core regtest nodes and how to resolve them quickly.

---

## 1. Zero Spendable Balance / Coinbase Maturity

### Symptom
You mined blocks to an address using `generatetoaddress`, but `getbalance` still returns `0.00000000` or transactions fail with:
```
Insufficient funds
```

### Cause
Bitcoin's `COINBASE_MATURITY` consensus rule means **newly mined coins (block subsidy + transaction fees) cannot be spent until they are 100 blocks deep** (so block N's reward is spendable once the chain reaches height N+100). This prevents reorganization race conditions where spent coinbase outputs vanish if an alternate chain tip wins.

### Solution
In regtest mode, always mine **at least 101 blocks** to mature the initial coinbase reward:
```bash
btc generatetoaddress 101 $(btc getnewaddress)
```
After 101 blocks, the reward from block 1 is 100 blocks deep and spendable (50 BTC). Mining 110 blocks gives you 10 matured rewards (500 BTC).

---

## 2. "Wallet file not specified" or "Requested wallet does not exist"

### Symptom
Running RPC commands like `btc getbalance` or `btc listunspent` produces:
```json
{
  "code": -19,
  "message": "Wallet file not specified (must request wallet RPC through /wallet/<filename> uri-path)"
}
```
or
```json
{
  "code": -18,
  "message": "Requested wallet does not exist or is not loaded"
}
```

### Cause
Bitcoin Core supports multi-wallet operation. In modern Bitcoin Core (v21+), wallets must be explicitly created or loaded into memory before wallet-specific RPCs can target them. If multiple wallets are loaded, the RPC client must route commands through `/wallet/<wallet_name>`.

### Solution
1. **List currently loaded wallets**:
   ```bash
   btc listwallets
   ```
2. **Create the default sandbox wallet if missing**:
   ```bash
   btc createwallet "lab"
   ```
3. **If already created but unloaded on restart**:
   ```bash
   btc loadwallet "lab"
   ```
4. **Using `bitcoin-cli` with a specific wallet**:
   ```bash
   btc -rpcwallet=lab getbalance
   ```
5. **Using JSON-RPC over HTTP**:
   Send HTTP POST to:
   ```
   http://127.0.0.1:18443/wallet/lab
   ```

---

## 3. ZeroMQ (ZMQ) Listener Receives No Events

### Symptom
Your subscriber script connects to `tcp://127.0.0.1:28332` (`rawblock`) or `tcp://127.0.0.1:28333` (`rawtx`), but `recv()` blocks indefinitely even when you run commands.

### Cause
Common causes include:
- ZeroMQ notifications only publish when an event *actually occurs* (e.g. a new block is mined or a new valid transaction enters the mempool). A passive node does not emit periodic heartbeat messages on ZMQ.
- ZeroMQ topic subscription was not registered or filtered out by topic prefix.
- ZeroMQ was not enabled in `bitcoin.conf` or ports differ from the subscriber settings.

### Solution
1. **Verify `bitcoin.conf` has ZMQ enabled**:
   ```ini
   zmqpubrawblock=tcp://0.0.0.0:28332
   zmqpubrawtx=tcp://0.0.0.0:28333
   ```
2. **Verify node listening sockets**:
   Inside the container / Linux environment:
   ```bash
   ss -tulpn | grep -E '28332|28333'
   ```
3. **Trigger events intentionally**:
   In a separate terminal, mine a block or send a transaction while the listener is running:
   ```bash
   btc generatetoaddress 1 $(btc getnewaddress)
   ```
4. **Check topic prefix subscription**:
   In ZeroMQ PUB/SUB, subscribers must subscribe to the topic string (e.g. `"rawblock"` or `"rawtx"`):
   ```python
   sock.setsockopt_string(zmq.SUBSCRIBE, "rawblock")
   ```

---

## 4. Port Forwarding & Host Authentication in Dev Containers

### Symptom
Connecting from your host machine (outside the Docker container) to `http://localhost:18443` fails with `Connection Refused` or `403 Forbidden`.

### Cause
- Bitcoin Core by default binds RPC to `127.0.0.1` inside the container network namespace, which prevents host port forwards from connecting unless `rpcbind=0.0.0.0` and `rpcallowip` permits access.
- Dev container ports were not forwarded in `devcontainer.json`.

### Solution
1. **Ensure `bitcoin.conf` allows container subnet connections**:
   ```ini
   rpcbind=0.0.0.0
   rpcallowip=0.0.0.0/0
   ```
   *(Note: Safe only in isolated regtest containers without public network exposure).*
2. **Ensure forwarded ports in `.devcontainer/devcontainer.json`**:
   Verify port `18443`, `28332`, `28333`, and `18444` are listed in `forwardPorts`.
3. **Always supply HTTP Basic Auth**:
   JSON-RPC requires authentication:
   ```bash
   curl --user bitcoinrpc:bitcoinrpcpassword \
        --data-binary '{"jsonrpc":"1.0","id":"test","method":"getblockchaininfo","params":[]}' \
        -H 'content-type: text/plain;' \
        http://127.0.0.1:18443/
   ```

---

## 5. Corrupted Regtest State or Node Won't Start

### Symptom
`bitcoind` fails to start with errors about database corruption, lock files, or dirty shutdowns:
```
Error: Cannot obtain a lock on data directory ...
```

### Cause
A previous node process did not shut down cleanly or another instance is already running.

### Solution
Use the provided reset script to wipe regtest chain data back to height 0 while preserving configuration:
```bash
bash scripts/reset-lab.sh
```
