/**
 * Lab 06: ZMQ Listener (TypeScript Reference Implementation)
 * Subscribe to Bitcoin Core's rawtx / rawblock ZeroMQ feeds, trigger one
 * transaction and one block, and verify the pushed bytes against the node.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";
import { getConfig } from "../../typescript/src/common/rpc.js";
import * as zmq from "zeromq";

const RECV_TIMEOUT_MS = 15_000;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg);
}

async function main() {
  console.log("=== Lab 06: ZMQ Listener (TypeScript) ===");
  const subTx = new zmq.Subscriber();
  const subBlock = new zmq.Subscriber();
  try {
    // Step 1
    process.stdout.write("[Step 1] Bootstrapping lab wallet & funds ... ");
    const { rpc: nodeRpc, walletRpc } = await bootstrapLab();
    const cfg = getConfig();
    console.log("✓");

    // Step 2
    process.stdout.write(
      `[Step 2] Subscribing to rawtx (${cfg.zmqRawTx}) and rawblock (${cfg.zmqRawBlock}) ... `,
    );
    subTx.receiveTimeout = RECV_TIMEOUT_MS;
    subTx.connect(cfg.zmqRawTx);
    subTx.subscribe("rawtx");
    subBlock.receiveTimeout = RECV_TIMEOUT_MS;
    subBlock.connect(cfg.zmqRawBlock);
    subBlock.subscribe("rawblock");
    await sleep(500); // let the SUB subscriptions propagate
    console.log("✓");

    // Step 3
    process.stdout.write("[Step 3] Broadcasting a transaction ... ");
    const dest = await walletRpc.call<string>("getnewaddress", ["zmq_probe", "bech32m"]);
    const sentTxid = await walletRpc.call<string>("sendtoaddress", [dest, 0.01]);
    console.log(`✓ (txid: ${sentTxid.slice(0, 16)}…)`);

    // Step 4
    process.stdout.write("[Step 4] Received rawtx frame & verified txid ... ");
    let frames: Buffer[];
    try {
      frames = (await subTx.receive()) as Buffer[];
    } catch {
      throw new Error(
        `no rawtx notification within ${RECV_TIMEOUT_MS} ms — is zmqpubrawtx set? (see docs/troubleshooting.md)`,
      );
    }
    const [topicTx, bodyTx, seqTx] = frames;
    assert(topicTx.toString() === "rawtx", `unexpected topic ${topicTx.toString()}`);
    const decoded = await nodeRpc.call<{ txid: string }>("decoderawtransaction", [bodyTx.toString("hex")]);
    assert(decoded.txid === sentTxid, `rawtx txid ${decoded.txid} != ${sentTxid}`);
    console.log(`✓ (seq ${seqTx.readUInt32LE(0)})`);

    // Step 5
    process.stdout.write("[Step 5] Mining 1 block ... ");
    const miner = await walletRpc.call<string>("getnewaddress", ["lab06_miner", "bech32m"]);
    const [blockHash] = await walletRpc.call<string[]>("generatetoaddress", [1, miner]);
    console.log(`✓ (hash: ${blockHash.slice(0, 16)}…)`);

    // Step 6
    process.stdout.write("[Step 6] Received rawblock frame & verified against getblock ... ");
    let bframes: Buffer[];
    try {
      bframes = (await subBlock.receive()) as Buffer[];
    } catch {
      throw new Error(
        `no rawblock notification within ${RECV_TIMEOUT_MS} ms — is zmqpubrawblock set? (see docs/troubleshooting.md)`,
      );
    }
    const [topicB, bodyB, seqB] = bframes;
    assert(topicB.toString() === "rawblock", `unexpected topic ${topicB.toString()}`);
    const expected = await nodeRpc.call<string>("getblock", [blockHash, 0]);
    assert(bodyB.toString("hex") === expected, "rawblock payload does not match getblock");
    console.log(`✓ (seq ${seqB.readUInt32LE(0)})`);

    console.log("======================================================");
    console.log("Result: PASS ✓");
    console.log("======================================================");
    process.exit(0);
  } catch (err: any) {
    console.error(`\n✗ FAILURE: ${err.message || err}`);
    console.log("======================================================");
    console.log("Result: FAIL ✗");
    console.log("======================================================");
    process.exit(1);
  } finally {
    subTx.close();
    subBlock.close();
  }
}

main();
