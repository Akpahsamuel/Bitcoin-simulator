/**
 * Lab 04: Multisig Escrow (TypeScript Reference Implementation)
 * 2-of-3 P2WSH built from raw script, funded from the lab wallet, then spent by
 * assembling the witness stack [<empty>, sigA, sigB, witnessScript] by hand.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";
import * as bitcoin from "bitcoinjs-lib";
import * as ecc from "@bitcoinerlab/secp256k1";

bitcoin.initEccLib(ecc);
const network = bitcoin.networks.regtest;

const SK_HEX = ["11".repeat(32), "22".repeat(32), "33".repeat(32)];
const EXPECT_WITNESS_SCRIPT =
  "5221034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa" +
  "2102466d7fcae563e5cb09a0d1870bb580344804617879a14949cf22285f1bae3f27" +
  "21023c72addb4fdf09af94f0c94d7fe92a386a7e70cf8a1d85916386bb2535c7b1b153ae";
const EXPECT_P2WSH = "bcrt1qpy8yjjs2l5neewx722mxve9w6m77zqsu7rldukggseflhwralerqh6ma0d";

const FUND_BTC = 0.5;
const FEE_SAT = 20_000;

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg);
}

async function main() {
  console.log("=== Lab 04: Multisig Escrow (TypeScript) ===");
  try {
    // Step 1
    process.stdout.write("[Step 1] Bootstrapping lab wallet & funds ... ");
    const { rpc: nodeRpc, walletRpc } = await bootstrapLab();
    console.log("✓");

    // Step 2
    process.stdout.write("[Step 2] Deriving 3 escrow keypairs from test vectors ... ");
    const privs = SK_HEX.map((h) => Buffer.from(h, "hex"));
    const pubs = privs.map((sk) => Buffer.from(ecc.pointFromScalar(sk, true)!));
    console.log(`✓ (${pubs.map((p) => p.toString("hex").slice(0, 10) + "…").join(", ")})`);

    // Step 3
    process.stdout.write("[Step 3] Building 2-of-3 witness script ... ");
    const p2ms = bitcoin.payments.p2ms({ m: 2, pubkeys: pubs, network });
    const witnessScript = p2ms.output!;
    assert(witnessScript.toString("hex") === EXPECT_WITNESS_SCRIPT, "witness script mismatch");
    console.log("✓ (matches canonical)");

    // Step 4
    process.stdout.write("[Step 4] Deriving P2WSH address & validating with node ... ");
    const p2wsh = bitcoin.payments.p2wsh({ redeem: p2ms, network });
    const addr = p2wsh.address!;
    assert(addr === EXPECT_P2WSH, `address ${addr} != ${EXPECT_P2WSH}`);
    const val = await nodeRpc.call<{ isvalid: boolean }>("validateaddress", [addr]);
    assert(val.isvalid, "node rejected the P2WSH address");
    console.log(`✓ (${addr.slice(0, 14)}…)`);

    // Step 5
    process.stdout.write(`[Step 5] Funding the multisig (${FUND_BTC} BTC) & mining 1 block ... `);
    const fundTxid = await walletRpc.call<string>("sendtoaddress", [addr, FUND_BTC]);
    const miner = await walletRpc.call<string>("getnewaddress", ["lab04_miner", "bech32m"]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    console.log(`✓ (txid: ${fundTxid.slice(0, 16)}…)`);

    // Step 6
    process.stdout.write("[Step 6] Locating the funding UTXO by scriptPubKey ... ");
    const spkHex = p2wsh.output!.toString("hex");
    const fundTx = await nodeRpc.call<any>("getrawtransaction", [fundTxid, true]);
    const out = fundTx.vout.find((o: any) => o.scriptPubKey.hex === spkHex);
    if (!out) throw new Error("funding output not found in transaction");
    const fundVout: number = out.n;
    const fundSat = Math.round(out.value * 1e8);
    console.log(`✓ (vout: ${fundVout}, ${out.value} BTC)`);

    // Step 7
    process.stdout.write("[Step 7] Building the spend transaction ... ");
    const returnAddr = await walletRpc.call<string>("getnewaddress", ["escrow_return", "bech32m"]);
    const outSat = fundSat - FEE_SAT;
    if (outSat <= 0) throw new Error("funded amount too small for fee");
    const tx = new bitcoin.Transaction();
    tx.version = 2;
    tx.addInput(Buffer.from(fundTxid, "hex").reverse(), fundVout);
    tx.addOutput(bitcoin.address.toOutputScript(returnAddr, network), outSat);
    console.log("✓");

    // Step 8
    process.stdout.write("[Step 8] BIP143 sighash + signing with keys A and B ... ");
    const sighash = tx.hashForWitnessV0(0, witnessScript, fundSat, bitcoin.Transaction.SIGHASH_ALL);
    const encode = (sk: Buffer) =>
      bitcoin.script.signature.encode(
        Buffer.from(ecc.sign(sighash, sk)),
        bitcoin.Transaction.SIGHASH_ALL,
      );
    const sigA = encode(privs[0]);
    const sigB = encode(privs[1]);
    console.log(`✓ (sigA ${sigA.length}B, sigB ${sigB.length}B)`);

    // Step 9
    process.stdout.write("[Step 9] Assembling witness [<empty>, sigA, sigB, script] ... ");
    tx.setWitness(0, [Buffer.alloc(0), sigA, sigB, witnessScript]);
    const rawHex = tx.toHex();
    console.log(`✓ (tx ${rawHex.length / 2} bytes)`);

    // Step 10
    process.stdout.write("[Step 10] testmempoolaccept ... ");
    const accept = await nodeRpc.call<any[]>("testmempoolaccept", [[rawHex]]);
    if (!accept?.[0]?.allowed) throw new Error(`mempool rejected: ${JSON.stringify(accept)}`);
    console.log("✓ (allowed: true)");

    // Step 11
    process.stdout.write("[Step 11] Broadcasting & mining 1 block ... ");
    const spendTxid = await nodeRpc.call<string>("sendrawtransaction", [rawHex]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    const details = await nodeRpc.call<{ confirmations?: number }>("getrawtransaction", [spendTxid, true]);
    if (!details.confirmations || details.confirmations < 1) {
      throw new Error(`expected confirmations >= 1, got ${details.confirmations}`);
    }
    console.log(`✓ (confirmations: ${details.confirmations})`);

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
