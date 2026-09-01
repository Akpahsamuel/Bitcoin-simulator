# Adding a Language to the Bitcoin Developer Sandbox

The **Bitcoin Developer Sandbox & Starter Kit** is designed to be language-agnostic. Any language implementation that adheres to the **Language Port Contract** can be integrated into the test runner and project suite.

This guide details the contract, standard directory conventions, and step-by-step instructions for adding a new language or porting an existing lab.

---

## 1. The Language Port Contract

Every language implementation of a lab **MUST** fulfill the following six invariants:

1. **Read Configuration Exclusively from `.env`**:
   - Connection settings, credentials, and wallet names must be loaded from `.env` (or environment variables).
   - Never hardcode ports (`18443`, `28332`, etc.), credentials (`bitcoinrpc`), URLs, or wallet names in example code.
   - The standard keys defined in [`.env.example`](../.env.example) are:
     ```env
     BITCOIN_RPC_URL=http://127.0.0.1:18443
     BITCOIN_RPC_USER=bitcoinrpc
     BITCOIN_RPC_PASSWORD=bitcoinrpcpassword
     BITCOIN_RPC_WALLET=lab
     BITCOIN_ZMQ_RAWBLOCK=tcp://127.0.0.1:28332
     BITCOIN_ZMQ_RAWTX=tcp://127.0.0.1:28333
     ```

2. **Call the Shared Bootstrap**:
   - Every example must be self-sufficient and idempotent.
   - At startup, call a language-specific bootstrap helper that checks if the `lab` wallet is loaded (or creates it) and verifies spendable balance. If balance is `0`, generate `101` blocks (`generatetoaddress 101 <address>`) so coinbase rewards mature (100 confirmations).

3. **Structured Step Output**:
   - Output clear, numbered steps indicating current operations with progress markers:
     - `[Step 1] Connect to node ... ✓`
     - `[Step 2] Query blockchain info ... ✓`
   - Finish with an explicit summary line: `PASS` or `ALL STEPS PASSED ✓`.

4. **Deterministic Exit Codes**:
   - Exit code **`0`** on complete success.
   - Exit code **non-zero (e.g., `1`)** immediately upon any assertion failure or unhandled exception.

5. **Single-Command Execution**:
   - Each example must be executable via a standard, single command from its directory (e.g., `python main.py`, `npm start` / `npx tsx main.ts`, `cargo run`, `go run .`).

6. **Offline / Regtest-Only Execution**:
   - The code must not require any external internet connectivity or reach out to mainnet, testnet, or signet.

---

## 2. Directory Layout & Naming Conventions

All examples follow a uniform directory structure:

```text
examples/
  <lang>/                      # Language-level shared packages/modules
    common/                    # Shared RPC client, bootstrap, utilities
  <NN-name>/                   # Numbered lab directory (e.g. 01-rpc-client)
    README.md                  # Goal, upstream book chapters, run commands
    python/main.py             # Python entry point
    typescript/main.ts         # TypeScript entry point
    rust/                      # Rust crate
      Cargo.toml
      src/main.rs
    go/                        # Go package
      go.mod
      main.go
    <new-lang>/                # Your new language port
      main.<ext>
```

### Key Rules:
- The entry point in every language directory MUST be named **`main.*`** (e.g., `main.py`, `main.ts`, `src/main.rs`, `main.go`, `main.rb`, `main.cs`).
- Version dependencies must be strictly pinned in the package manifest (`requirements.txt`, `package.json`, `Cargo.toml`, `go.mod`, etc.).

---

## 3. Step-by-Step: Porting or Adding a New Language

### Step 3.1: Create the Shared Language Directory
Create `examples/<lang>/` to host common utilities:
1. An RPC helper that parses `.env` and issues HTTP JSON-RPC calls (with Basic Authentication) and REST queries.
2. A `bootstrap` helper that loads or creates the `lab` wallet and mines 101 blocks to achieve spendable funds.

### Step 3.2: Implement Lab 01 (`01-rpc-client`)
1. Create directory `examples/01-rpc-client/<lang>/`.
2. Add your entry file `main.<ext>`.
3. Implement the lab steps:
   - Authenticate with the node via JSON-RPC.
   - Call `getblockchaininfo`.
   - Call `getnewaddress` and `generatetoaddress 1 <addr>`.
   - Query balance via `getbalance`.
   - Call unauthenticated REST endpoint `GET /rest/chaininfo.json`.
4. Ensure each step prints `✓` and exits `0` on success.

### Step 3.3: Implement Labs 02 & 03
1. `02-keys-and-addresses`: Derive BIP39 mnemonics, BIP32/BIP44 paths, BIP84 SegWit, and BIP86 Taproot addresses, validating against node with `validateaddress`.
2. `03-raw-transactions`: Query UTXOs with `listunspent`, compute transaction vsize & sat/vB fees, build raw transaction with `createrawtransaction`, sign with wallet (`signrawtransactionwithwallet`), verify with `testmempoolaccept`, broadcast with `sendrawtransaction`, and confirm via mining.

### Step 3.4: Integrate with the Test Runner
Update [`scripts/test-examples.sh`](../scripts/test-examples.sh) to add execution support for `<lang>`:
```bash
<new-lang>)
    (cd "$lang_dir" && <run_command>) > /tmp/test_output.log 2>&1 || run_status=$?
    ;;
```

### Step 3.5: Update Documentation & Dockerfile
- Add the language runtime and build dependencies to [`.devcontainer/Dockerfile`](../.devcontainer/Dockerfile).
- Document run commands in each lab's `README.md`.
- Run `bash scripts/test-examples.sh` to confirm all tests execute and pass cleanly.
