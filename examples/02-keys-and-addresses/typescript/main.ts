/**
 * Lab 02: Keys and Addresses (TypeScript Reference Implementation)
 * Demonstrates BIP39 mnemonics, BIP32/BIP44 derivation, Legacy/SegWit/Taproot addresses,
 * and node consensus validation.
 */
import { bootstrapLab } from "../../typescript/src/common/bootstrap.js";
import * as bip39 from "bip39";
import { BIP32Factory } from "bip32";
import * as ecc from "@bitcoinerlab/secp256k1";
import * as bitcoin from "bitcoinjs-lib";

// Initialize ECC library for Taproot and BIP32
bitcoin.initEccLib(ecc);
const bip32 = BIP32Factory(ecc);
const network = bitcoin.networks.regtest;

// Deterministic test vector: the canonical "abandon ... about" mnemonic on regtest.
// Expected values are cross-checked against embit (Python), rust-bitcoin (Rust) and
// btcd (Go) — every port must derive exactly these.
const TEST_MNEMONIC =
  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";
const EXPECT_FINGERPRINT = "73c5da0a";
const EXPECT_P2PKH = "mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV";
const EXPECT_P2WPKH = "bcrt1q6rz28mcfaxtmd6v789l9rrlrusdprr9pz3cppk";
const EXPECT_P2TR = "bcrt1p8wpt9v4frpf3tkn0srd97pksgsxc5hs52lafxwru9kgeephvs7rqjeprhg";

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg);
}

async function main() {
  console.log("=== Lab 02: Keys & Addresses (TypeScript) ===");

  try {
    // Step 1: Bootstrap lab
    process.stdout.write("[Step 1] Bootstrapping lab wallet & RPC connection ... ");
    const { rpc: nodeRpc } = await bootstrapLab();
    console.log("✓");

    // Step 2: BIP39 mnemonic -> 512-bit root seed
    process.stdout.write("[Step 2] Generating BIP39 mnemonic & root seed ... ");
    const seed = await bip39.mnemonicToSeed(TEST_MNEMONIC);
    assert(seed.length === 64, `Expected 64-byte seed, got ${seed.length}`);
    console.log(`✓ (12 words, ${seed.length}-byte seed)`);

    // Step 3: BIP32 Root Key
    process.stdout.write("[Step 3] Deriving BIP32 root key ... ");
    const root = bip32.fromSeed(seed, network);
    const fingerprint = Buffer.from(root.fingerprint).toString("hex");
    assert(fingerprint === EXPECT_FINGERPRINT, `fingerprint ${fingerprint} != ${EXPECT_FINGERPRINT}`);
    console.log(`✓ (fingerprint: ${fingerprint})`);

    // Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
    process.stdout.write("[Step 4] Deriving BIP44 Legacy P2PKH address ... ");
    const childP2PKH = root.derivePath("m/44'/1'/0'/0/0");
    const { address: addrP2PKH } = bitcoin.payments.p2pkh({
      pubkey: Buffer.from(childP2PKH.publicKey),
      network,
    });
    assert(addrP2PKH === EXPECT_P2PKH, `P2PKH ${addrP2PKH} != ${EXPECT_P2PKH}`);
    console.log(`✓ (${addrP2PKH})`);

    // Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
    process.stdout.write("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ");
    const childP2WPKH = root.derivePath("m/84'/1'/0'/0/0");
    const { address: addrP2WPKH } = bitcoin.payments.p2wpkh({
      pubkey: Buffer.from(childP2WPKH.publicKey),
      network,
    });
    assert(addrP2WPKH === EXPECT_P2WPKH, `P2WPKH ${addrP2WPKH} != ${EXPECT_P2WPKH}`);
    console.log(`✓ (${addrP2WPKH})`);

    // Step 6: Taproot P2TR (m/86'/1'/0'/0/0, BIP86 key-path tweak applied by p2tr)
    process.stdout.write("[Step 6] Deriving BIP86 Taproot P2TR address ... ");
    const childP2TR = root.derivePath("m/86'/1'/0'/0/0");
    const internalPubkey = Buffer.from(childP2TR.publicKey).subarray(1, 33);
    const { address: addrP2TR } = bitcoin.payments.p2tr({
      internalPubkey,
      network,
    });
    assert(addrP2TR === EXPECT_P2TR, `P2TR ${addrP2TR} != ${EXPECT_P2TR}`);
    console.log(`✓ (${addrP2TR})`);

    // Step 7: Validate derived addresses with Bitcoin Core node
    process.stdout.write("[Step 7] Validating addresses against Bitcoin Core node ... ");
    for (const addr of [addrP2PKH, addrP2WPKH, addrP2TR]) {
      if (!addr) throw new Error("Derived empty address");
      const val = await nodeRpc.call<{ isvalid: boolean }>("validateaddress", [addr]);
      if (!val.isvalid) {
        throw new Error(`Address ${addr} is reported invalid by node`);
      }
    }
    console.log("✓ (all addresses valid)");

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
