# Architecture Deep Dive

How the sandbox's regtest node, its three interfaces (RPC, REST, ZMQ), and the
four language clients fit together. This is a reference for understanding the
*mechanics*, not a getting-started guide — start with
[`examples/00-cli-workshop/`](../examples/00-cli-workshop/README.md) if you
haven't run the lab yet, and see [`troubleshooting.md`](troubleshooting.md) if
something is failing.

---

## 1. Why regtest

Bitcoin Core ships four networks: mainnet, testnet, signet, and **regtest**
(regression test). Only regtest gives you:

- **On-demand block generation.** `generatetoaddress` mines a block
  immediately at minimal difficulty — no hashpower, no waiting on a public
  network's block time. Every lab controls exactly when blocks appear.
- **A private, disposable chain.** The node has zero peers by default; the
  entire chain state lives under `$HOME/btc-lab` and `reset-lab.sh` deletes it
  back to height 0. Nothing you do here touches real money or a shared
  ledger (see CLAUDE.md golden rule #2 — no code path may reach mainnet,
  testnet, or signet).
- **Deterministic parameters.** Regtest disables retargeting and lowers
  standardness/relay rules just enough that hand-built scripts (multisig,
  timelocks) behave predictably across runs.

The one consensus rule regtest does **not** relax is **coinbase maturity**:
newly mined block rewards are unspendable for 100 blocks, same as mainnet.
Every bootstrap helper mines 101 blocks specifically to clear this — see
`docs/troubleshooting.md` §1 for the mechanics.

---

## 2. Process & data directory

`bitcoind` is a single long-running process per lab session, started by
`scripts/init-lab.sh` with `-datadir="$HOME/btc-lab"` (override via
`BTC_DATADIR`). That directory holds:

- `bitcoin.conf` — written by `init-lab.sh`, never hand-edited. It carries a
  `rev` marker in a comment; a stale layout from an older sandbox version is
  detected and rewritten (the previous copy saved as `bitcoin.conf.bak`).
- `regtest/` — the chain state: block index, chainstate (UTXO set), mempool
  dump, and the `wallets/lab/` descriptor wallet directory.
- `debug.log` — bitcoind's own log; check here first when an RPC call hangs
  or a ZMQ frame never arrives.

### Global vs. `[regtest]` config sections

```
regtest=1
server=1
rest=1
txindex=1
rpcuser=bitcoinrpc
rpcpassword=bitcoinrpcpassword

[regtest]
fallbackfee=0.0002
rpcbind=0.0.0.0
rpcallowip=0.0.0.0/0
zmqpubrawblock=tcp://0.0.0.0:28332
zmqpubrawtx=tcp://0.0.0.0:28333
```

Bitcoin Core enforces that **network-specific** options — anything that binds
a socket or sets a network-dependent policy — be sectioned under
`[regtest]`; putting them in the global section makes `bitcoind` refuse to
start on regtest. `rpcbind=0.0.0.0` / `rpcallowip=0.0.0.0/0` look permissive
but are safe here: the node is only ever reachable via localhost port-forwards
from the devcontainer host, never a public interface.

### Descriptor wallets only

Core 31 ships **no legacy Berkeley-DB wallet support** — every wallet created
by this sandbox is a descriptor wallet (`createwallet` with no BDB flags).
That means RPCs which mutated a legacy wallet's keypool directly
(`importprivkey`, `importaddress`, `addmultisigaddress`, `dumpprivkey`,
`sethdseed`) are simply absent. Anything resembling multisig or a custom
script (labs `04`–`05`) is therefore built **in the example's own crypto
library** (`embit` / `bitcoinjs-lib` / `rust-bitcoin` / `btcsuite`) and only
touches the node wallet to fund an address and broadcast the finished,
already-signed transaction. Watch-only tracking of a hand-built address uses
`importdescriptors`, the descriptor-wallet-native replacement for
`importaddress`.

---

## 3. JSON-RPC interface (port 18443)

A single HTTP endpoint, `POST http://127.0.0.1:18443/`, speaking **JSON-RPC
1.0** with HTTP Basic auth (`bitcoinrpc` / `bitcoinrpcpassword` — intentionally
public, regtest-only credentials, never wired to anything that could reach a
real network):

```json
// request
{"jsonrpc": "1.0", "id": "sandbox", "method": "getblockchaininfo", "params": []}
// response
{"result": { "chain": "regtest", "blocks": 0, ... }, "error": null, "id": "sandbox"}
```

A non-null `error` object (`{"code": ..., "message": ...}`) always
accompanies HTTP 200 or 500 — never trust the HTTP status alone; check
`error` first.

### Wallet routing

Bitcoin Core supports multiple loaded wallets simultaneously, so
wallet-scoped RPCs (`getbalance`, `getnewaddress`, `listunspent`,
`sendtoaddress`, `signrawtransactionwithwallet`, …) must be routed through
`/wallet/<name>` rather than `/`:

```
POST http://127.0.0.1:18443/wallet/lab
```

Node-level RPCs (`getblockchaininfo`, `generatetoaddress`, `decoderawtransaction`,
`getblock`, `sendrawtransaction`, `testmempoolaccept`) work against the bare
`/` endpoint and take no wallet context — they operate on chain state, not
wallet state. This is why every language's `common/` client exposes both an
unscoped RPC handle and a `for_wallet("lab")` / wallet-constructed variant
(see [`cheatsheet.md`](cheatsheet.md) for the exact call in each language).

---

## 4. REST interface (same port, `/rest/` path)

Read-only chain queries, **no authentication**, `GET` only, mounted under
`/rest/` on the same port as RPC. REST exists precisely because it needs no
credentials: it exposes only public, already-consensus-final data (blocks,
transactions, UTXO set, mempool contents), so a block explorer or indexer can
be handed a REST URL without ever seeing the RPC password. Contrast with RPC,
which can mutate wallet and node state and therefore must stay
authenticated.

Every REST path takes a format suffix — `.json` for structured data, `.bin`
for the raw serialized bytes, `.hex` for the same bytes as a hex string. Labs
in this sandbox use `.json` exclusively.

---

## 5. ZMQ interface (ports 28332 / 28333)

Bitcoin Core can publish node events over ZeroMQ PUB sockets. This sandbox's
`bitcoin.conf` enables exactly two of Core's five possible topics:

| Topic | Port | Payload |
| :--- | :--- | :--- |
| `rawblock` | 28332 | Full serialized block bytes |
| `rawtx` | 28333 | Full serialized transaction bytes |

(Core also supports `hashblock`, `hashtx`, and `sequence` topics — this
sandbox does not enable them, since every lab needs the full bytes to decode
with `decoderawtransaction` / feed to `getblock`, not just a hash.)

Each socket is standard ZMQ PUB/SUB: a subscriber connects, subscribes to a
topic prefix, and receives a **3-frame multipart message**:

```
[ topic:  b"rawtx" or b"rawblock",
  body:   raw bytes (the tx or block, exactly as node-serialized),
  sequence: 4-byte little-endian uint32, increments per topic ]
```

The sequence number is the only ordering/loss signal ZMQ gives you — PUB/SUB
has no replay and no history, so **a subscriber must connect and subscribe
before the event happens**, or the notification is gone forever. This is the
fundamental difference from RPC: RPC is pull (ask "what's the mempool now?"
any time), ZMQ is push (be listening at the moment something changes). Lab
`06-zmq-listener` demonstrates the pattern every language uses: subscribe
first, give the socket a moment to propagate the subscription, *then*
trigger the transaction/block, then read.

---

## 6. Port & interface summary

| Interface | Address | Auth | Direction |
| :--- | :--- | :--- | :--- |
| JSON-RPC | `http://127.0.0.1:18443/` | HTTP Basic | Request/response |
| REST | `http://127.0.0.1:18443/rest/` | None | Request/response (GET only) |
| ZMQ rawblock | `tcp://127.0.0.1:28332` | None | Push (PUB/SUB) |
| ZMQ rawtx | `tcp://127.0.0.1:28333` | None | Push (PUB/SUB) |
| P2P | `127.0.0.1:18444` | None | Bitcoin wire protocol (unused by the labs) |

---

## 7. How the four language clients fit in

Every lab's `<lang>/main.*` is a thin script; the interface logic lives once
per language in a shared `common/` module, so the wire-level details above are
implemented exactly once per language, not once per lab:

| Language | RPC + REST client | Bootstrap (wallet + maturity) | ZMQ (lab `06` only) |
| :--- | :--- | :--- | :--- |
| Python | `examples/python/common/rpc.py` | `examples/python/common/bootstrap.py` | `pyzmq` |
| TypeScript | `examples/typescript/src/common/rpc.ts` | `examples/typescript/src/common/bootstrap.ts` | `zeromq` (npm) |
| Rust | `examples/rust/common/src/rpc.rs` | `examples/rust/common/src/bootstrap.rs` | `zmq` crate (sync) |
| Go | `examples/go/common/rpc.go` | `examples/go/common/bootstrap.go` | `pebbe/zmq4` |

Each client reads connection settings from the repo-root `.env` (walking
parent directories to find it — labs run from their own `<NN>/<lang>/`
subdirectory), never hardcodes a port, credential, or wallet name, and exposes
the same shape: a `call(method, params)` for RPC, a REST getter, and a
wallet-scoped variant. See [`cheatsheet.md`](cheatsheet.md) for the concrete
call in each language.
