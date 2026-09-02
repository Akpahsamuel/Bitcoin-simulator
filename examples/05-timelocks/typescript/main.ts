/**
 * Lab 05: Timelocks (TypeScript Reference Implementation)
 * CLTV (absolute, BIP65) and CSV (relative, BIP112) P2WSH outputs: build the
 * script, fund it, watch testmempoolaccept reject an early spend, advance the
 * chain, then spend successfully.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";
import type { BitcoinRPC } from "../../typescript/src/common/rpc.js";
import * as bitcoin from "bitcoinjs-lib";
import * as ecc from "@bitcoinerlab/secp256k1";

bitcoin.initEccLib(ecc);
const network = bitcoin.networks.regtest;
const { opcodes, script: bscript, Transaction } = bitcoin;

const SK_A_HEX = "11".repeat(32);
const EXPECT_CSV_SCRIPT =
  "53b27521034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aaac";
const EXPECT_CSV_ADDR = "bcrt1ql739fkda7sf20qkdwgku2j0ppeff4r7vsqasvvxestsqwtvuak3s9rktmg";

const CSV_DELAY = 3;
const FUND_BTC = 0.2;
const FEE_SAT = 20_000;

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg);
}

const p2wshAddr = (witnessScript: Buffer) =>
  bitcoin.payments.p2wsh({ redeem: { output: witnessScript, network }, network });

async function findVout(rpc: BitcoinRPC, txid: string, spkHex: string) {
  const tx = await rpc.call<any>("getrawtransaction", [txid, true]);
  const o = tx.vout.find((v: any) => v.scriptPubKey.hex === spkHex);
  if (!o) throw new Error("funding output not found");
  return { vout: o.n as number, sat: Math.round(o.value * 1e8) };
}

async function buildSpend(
  rpc: BitcoinRPC,
  skA: Buffer,
  witnessScript: Buffer,
  txid: string,
  vout: number,
  valueSat: number,
  opts: { version: number; sequence: number; locktime: number },
): Promise<string> {
  const dest = await rpc.call<string>("getnewaddress", ["timelock_return", "bech32m"]);
  const outSat = valueSat - FEE_SAT;
  if (outSat <= 0) throw new Error("funded amount too small for fee");
  const tx = new Transaction();
  tx.version = opts.version;
  tx.locktime = opts.locktime;
  tx.addInput(Buffer.from(txid, "hex").reverse(), vout, opts.sequence);
  tx.addOutput(bitcoin.address.toOutputScript(dest, network), outSat);
  const sighash = tx.hashForWitnessV0(0, witnessScript, valueSat, Transaction.SIGHASH_ALL);
  const sig = bscript.signature.encode(Buffer.from(ecc.sign(sighash, skA)), Transaction.SIGHASH_ALL);
  tx.setWitness(0, [sig, witnessScript]);
  return tx.toHex();
}

async function main() {
  console.log("=== Lab 05: Timelocks (TypeScript) ===");
  try {
    process.stdout.write("[Step 1] Bootstrapping lab wallet & funds ... ");
    const { rpc: nodeRpc, walletRpc } = await bootstrapLab();
    const skA = Buffer.from(SK_A_HEX, "hex");
    const pubA = Buffer.from(ecc.pointFromScalar(skA, true)!);
    const miner = await walletRpc.call<string>("getnewaddress", ["lab05_miner", "bech32m"]);
    console.log("✓");

    // ---------- CLTV ----------
    process.stdout.write("[Step 2] CLTV: building <height> OP_CLTV script ... ");
    const lockHeight = (await nodeRpc.call<number>("getblockcount")) + 10;
    const cltvScript = bscript.compile([
      bscript.number.encode(lockHeight),
      opcodes.OP_CHECKLOCKTIMEVERIFY,
      opcodes.OP_DROP,
      pubA,
      opcodes.OP_CHECKSIG,
    ]);
    const cltv = p2wshAddr(cltvScript);
    console.log(`✓ (lock at height ${lockHeight})`);

    process.stdout.write("[Step 3] CLTV: deriving P2WSH address & validating with node ... ");
    assert((await nodeRpc.call<{ isvalid: boolean }>("validateaddress", [cltv.address!])).isvalid, "node rejected CLTV address");
    console.log(`✓ (${cltv.address!.slice(0, 14)}…)`);

    process.stdout.write(`[Step 4] CLTV: funding (${FUND_BTC} BTC) & mining 1 block ... `);
    const cltvTxid = await walletRpc.call<string>("sendtoaddress", [cltv.address, FUND_BTC]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    const cltvUtxo = await findVout(nodeRpc, cltvTxid, cltv.output!.toString("hex"));
    console.log(`✓ (vout ${cltvUtxo.vout})`);

    process.stdout.write("[Step 5] CLTV: early spend rejected by testmempoolaccept ... ");
    let raw = await buildSpend(walletRpc, skA, cltvScript, cltvTxid, cltvUtxo.vout, cltvUtxo.sat, {
      version: 2,
      sequence: 0xfffffffe,
      locktime: lockHeight,
    });
    let res = await nodeRpc.call<any[]>("testmempoolaccept", [[raw]]);
    assert(res?.[0]?.allowed === false, `expected rejection, got ${JSON.stringify(res)}`);
    console.log(`✓ (allowed: false — ${res[0]["reject-reason"]})`);

    process.stdout.write("[Step 6] CLTV: mining to the lock height ... ");
    const need = lockHeight - (await nodeRpc.call<number>("getblockcount")) + 1;
    if (need > 0) await walletRpc.call("generatetoaddress", [need, miner]);
    console.log(`✓ (height ${await nodeRpc.call<number>("getblockcount")})`);

    process.stdout.write("[Step 7] CLTV: spend accepted, broadcast & confirmed ... ");
    raw = await buildSpend(walletRpc, skA, cltvScript, cltvTxid, cltvUtxo.vout, cltvUtxo.sat, {
      version: 2,
      sequence: 0xfffffffe,
      locktime: lockHeight,
    });
    let acc = await nodeRpc.call<any[]>("testmempoolaccept", [[raw]]);
    assert(acc?.[0]?.allowed === true, `mempool rejected: ${JSON.stringify(acc)}`);
    let spendTxid = await nodeRpc.call<string>("sendrawtransaction", [raw]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    let confs = (await nodeRpc.call<{ confirmations?: number }>("getrawtransaction", [spendTxid, true])).confirmations ?? 0;
    assert(confs >= 1, `confirmations ${confs}`);
    console.log(`✓ (confirmations: ${confs})`);

    // ---------- CSV ----------
    process.stdout.write("[Step 8] CSV: building <3> OP_CSV script ... ");
    const csvScript = bscript.compile([
      bscript.number.encode(CSV_DELAY),
      opcodes.OP_CHECKSEQUENCEVERIFY,
      opcodes.OP_DROP,
      pubA,
      opcodes.OP_CHECKSIG,
    ]);
    assert(csvScript.toString("hex") === EXPECT_CSV_SCRIPT, "CSV script mismatch");
    const csv = p2wshAddr(csvScript);
    console.log("✓ (matches canonical)");

    process.stdout.write("[Step 9] CSV: deriving P2WSH address & validating with node ... ");
    assert(csv.address === EXPECT_CSV_ADDR, `${csv.address} != ${EXPECT_CSV_ADDR}`);
    assert((await nodeRpc.call<{ isvalid: boolean }>("validateaddress", [csv.address!])).isvalid, "node rejected CSV address");
    console.log(`✓ (${csv.address!.slice(0, 14)}…)`);

    process.stdout.write(`[Step 10] CSV: funding (${FUND_BTC} BTC) & mining 1 block ... `);
    const csvTxid = await walletRpc.call<string>("sendtoaddress", [csv.address, FUND_BTC]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    const csvUtxo = await findVout(nodeRpc, csvTxid, csv.output!.toString("hex"));
    console.log(`✓ (vout ${csvUtxo.vout}, 1 confirmation)`);

    process.stdout.write("[Step 11] CSV: early spend rejected by testmempoolaccept ... ");
    raw = await buildSpend(walletRpc, skA, csvScript, csvTxid, csvUtxo.vout, csvUtxo.sat, {
      version: 2,
      sequence: CSV_DELAY,
      locktime: 0,
    });
    res = await nodeRpc.call<any[]>("testmempoolaccept", [[raw]]);
    assert(res?.[0]?.allowed === false, `expected rejection, got ${JSON.stringify(res)}`);
    console.log(`✓ (allowed: false — ${res[0]["reject-reason"]})`);

    process.stdout.write("[Step 12] CSV: mining 3 blocks to satisfy the relative delay ... ");
    await walletRpc.call("generatetoaddress", [CSV_DELAY, miner]);
    console.log("✓");

    process.stdout.write("[Step 13] CSV: spend accepted, broadcast & confirmed ... ");
    raw = await buildSpend(walletRpc, skA, csvScript, csvTxid, csvUtxo.vout, csvUtxo.sat, {
      version: 2,
      sequence: CSV_DELAY,
      locktime: 0,
    });
    acc = await nodeRpc.call<any[]>("testmempoolaccept", [[raw]]);
    assert(acc?.[0]?.allowed === true, `mempool rejected: ${JSON.stringify(acc)}`);
    spendTxid = await nodeRpc.call<string>("sendrawtransaction", [raw]);
    await walletRpc.call("generatetoaddress", [1, miner]);
    confs = (await nodeRpc.call<{ confirmations?: number }>("getrawtransaction", [spendTxid, true])).confirmations ?? 0;
    assert(confs >= 1, `confirmations ${confs}`);
    console.log(`✓ (confirmations: ${confs})`);

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
