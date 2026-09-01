/**
 * Lab 03: Raw Transactions (TypeScript Reference Implementation)
 * Demonstrates manual UTXO selection, fee calculation, raw transaction construction,
 * signing, mempool acceptance testing, and broadcast.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";

interface UTXO {
  txid: string;
  vout: number;
  amount: number;
}

interface SignResult {
  hex: string;
  complete: boolean;
}

interface MempoolAcceptResult {
  txid: string;
  allowed: boolean;
  rejectReason?: string;
}

async function main() {
  console.log("=== Lab 03: Raw Transactions (TypeScript) ===");

  try {
    // Step 1: Bootstrap lab
    process.stdout.write("[Step 1] Bootstrapping lab wallet & ensure spendable UTXOs ... ");
    const { rpc: nodeRpc, walletRpc } = await bootstrapLab();
    console.log("✓");

    // Step 2: Select UTXO from listunspent
    process.stdout.write("[Step 2] Selecting UTXO from listunspent ... ");
    const unspent = await walletRpc.call<UTXO[]>("listunspent", [1, 9999999]);
    if (!unspent || unspent.length === 0) {
      throw new Error("No spendable UTXOs found in wallet");
    }

    const utxo = unspent[0];
    console.log(`✓ (txid: ${utxo.txid.substring(0, 16)}..., vout: ${utxo.vout}, amount: ${utxo.amount} BTC)`);

    // Step 3: Construct raw transaction
    process.stdout.write("[Step 3] Constructing raw transaction hex (createrawtransaction) ... ");
    const recipientAddr = await walletRpc.call<string>("getnewaddress", ["recipient", "bech32m"]);
    const changeAddr = await walletRpc.call<string>("getnewaddress", ["change", "bech32m"]);

    const sendAmount = 1.5;
    const fee = 0.0001;
    const changeAmount = Number((utxo.amount - sendAmount - fee).toFixed(8));
    if (changeAmount <= 0) {
      throw new Error(`UTXO amount (${utxo.amount}) insufficient for send + fee`);
    }

    const inputs = [{ txid: utxo.txid, vout: utxo.vout }];
    const outputs = [
      { [recipientAddr]: sendAmount },
      { [changeAddr]: changeAmount },
    ];

    const rawTxHex = await walletRpc.call<string>("createrawtransaction", [inputs, outputs]);
    if (!rawTxHex) throw new Error("createrawtransaction failed");
    console.log(`✓ (hex length: ${rawTxHex.length})`);

    // Step 4: Sign transaction inputs
    process.stdout.write("[Step 4] Signing transaction inputs (signrawtransactionwithwallet) ... ");
    const signResult = await walletRpc.call<SignResult>("signrawtransactionwithwallet", [rawTxHex]);
    if (!signResult.complete) {
      throw new Error("Transaction signing was incomplete");
    }
    console.log("✓ (complete: true)");

    // Step 5: Verify transaction with testmempoolaccept
    process.stdout.write("[Step 5] Verifying transaction with testmempoolaccept ... ");
    const mempoolTest = await nodeRpc.call<MempoolAcceptResult[]>("testmempoolaccept", [[signResult.hex]]);
    if (!mempoolTest || !mempoolTest[0]?.allowed) {
      throw new Error(`Mempool reject: ${JSON.stringify(mempoolTest)}`);
    }
    console.log("✓ (allowed: true)");

    // Step 6: Broadcast transaction via sendrawtransaction
    process.stdout.write("[Step 6] Broadcasting transaction via sendrawtransaction ... ");
    const broadcastTxid = await nodeRpc.call<string>("sendrawtransaction", [signResult.hex]);
    if (!broadcastTxid || broadcastTxid.length !== 64) {
      throw new Error(`Invalid broadcast txid: ${broadcastTxid}`);
    }
    console.log(`✓ (txid: ${broadcastTxid})`);

    // Step 7: Mine 1 block & verify confirmation
    process.stdout.write("[Step 7] Mining 1 block & verifying confirmation ... ");
    const minerAddr = await walletRpc.call<string>("getnewaddress", ["miner", "bech32m"]);
    await walletRpc.call("generatetoaddress", [1, minerAddr]);

    const txDetails = await walletRpc.call<{ confirmations: number }>("getrawtransaction", [broadcastTxid, true]);
    if (!txDetails.confirmations || txDetails.confirmations < 1) {
      throw new Error(`Expected confirmations >= 1, got ${txDetails.confirmations}`);
    }
    console.log(`✓ (confirmations: ${txDetails.confirmations})`);

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
