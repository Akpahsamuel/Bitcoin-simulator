/**
 * Lab 01: RPC Client (TypeScript Reference Implementation)
 * Demonstrates JSON-RPC authentication, wallet operations, mining, and REST queries.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";

async function main() {
  console.log("=== Lab 01: RPC Client (TypeScript) ===");

  try {
    // Step 1: Bootstrap lab wallet & coinbase maturity
    process.stdout.write("[Step 1] Bootstrapping lab wallet and initial funds ... ");
    const { rpc: nodeRpc, walletRpc } = await bootstrapLab();
    console.log("✓");

    // Step 2: Query blockchain info via JSON-RPC
    process.stdout.write("[Step 2] Querying getblockchaininfo via JSON-RPC ... ");
    const chainInfo = await nodeRpc.call<{ chain: string; blocks: number }>("getblockchaininfo");
    if (chainInfo.chain !== "regtest") {
      throw new Error(`Expected regtest, got ${chainInfo.chain}`);
    }
    if (chainInfo.blocks < 101) {
      throw new Error(`Expected at least 101 blocks, got ${chainInfo.blocks}`);
    }
    console.log(`✓ (chain: ${chainInfo.chain}, blocks: ${chainInfo.blocks})`);

    // Step 3: Generate fresh address and mine 1 block
    process.stdout.write("[Step 3] Generating fresh address and mining 1 block ... ");
    const freshAddr = await walletRpc.call<string>("getnewaddress", ["lab01_test", "bech32m"]);
    const blockHashes = await walletRpc.call<string[]>("generatetoaddress", [1, freshAddr]);
    if (!blockHashes || blockHashes.length !== 1) {
      throw new Error("Expected 1 block mined");
    }
    const newBlocks = await nodeRpc.call<number>("getblockcount");
    console.log(`✓ (new height: ${newBlocks}, mined block: ${blockHashes[0].substring(0, 16)}...)`);

    // Step 4: Query wallet balance
    process.stdout.write("[Step 4] Querying wallet balance via getbalance ... ");
    const balance = await walletRpc.call<number>("getbalance");
    if (balance <= 0) {
      throw new Error(`Expected spendable balance > 0, got ${balance}`);
    }
    console.log(`✓ (balance: ${balance} BTC)`);

    // Step 5: Query unauthenticated REST API
    process.stdout.write("[Step 5] Querying unauthenticated REST API (/rest/chaininfo.json) ... ");
    const restInfo = await nodeRpc.getRest<{ chain: string; blocks: number }>("chaininfo.json");
    if (restInfo.chain !== "regtest" || restInfo.blocks !== newBlocks) {
      throw new Error("REST API verification failed");
    }
    console.log("✓");

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
  }
}

main();
