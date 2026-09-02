//! Lab 05: Timelocks (Rust Reference Implementation)
//! CLTV (absolute, BIP65) and CSV (relative, BIP112) P2WSH outputs: build the
//! script, fund it, watch testmempoolaccept reject an early spend, advance the
//! chain, then spend successfully.

use bitcoin::absolute::LockTime;
use bitcoin::consensus::encode::serialize_hex;
use bitcoin::hashes::Hash;
use bitcoin::opcodes::all::{OP_CHECKSIG, OP_CLTV, OP_CSV, OP_DROP};
use bitcoin::script::{Builder, ScriptBuf};
use bitcoin::secp256k1::{Message, Secp256k1, SecretKey};
use bitcoin::sighash::{EcdsaSighashType, SighashCache};
use bitcoin::transaction::Version;
use bitcoin::{
    Address, Amount, Network, OutPoint, PublicKey, Sequence, Transaction, TxIn, TxOut, Txid, Witness,
};
use bitcoin_sandbox_common::{bootstrap_lab, BitcoinRPC};
use serde_json::{json, Value};
use std::io::{self, Write};
use std::process;
use std::str::FromStr;

const SK_A_HEX: &str = "1111111111111111111111111111111111111111111111111111111111111111";
const EXPECT_CSV_SCRIPT: &str =
    "53b27521034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aaac";
const EXPECT_CSV_ADDR: &str = "bcrt1ql739fkda7sf20qkdwgku2j0ppeff4r7vsqasvvxestsqwtvuak3s9rktmg";

const CSV_DELAY: i64 = 3;
const FUND_BTC: f64 = 0.2;
const FEE_SAT: u64 = 20_000;

fn fail(msg: &str) -> ! {
    eprintln!("\n✗ FAILURE: {}", msg);
    println!("======================================================");
    println!("Result: FAIL ✗");
    println!("======================================================");
    process::exit(1);
}

macro_rules! step {
    ($($arg:tt)*) => {{ print!($($arg)*); io::stdout().flush().unwrap(); }};
}

fn rpc(r: &BitcoinRPC, method: &str, params: Vec<Value>) -> Value {
    r.call(method, params).unwrap_or_else(|e| fail(&e))
}

fn find_vout(node: &BitcoinRPC, txid: &str, spk_hex: &str) -> (u32, u64) {
    let tx = rpc(node, "getrawtransaction", vec![json!(txid), json!(true)]);
    let o = tx["vout"]
        .as_array()
        .and_then(|outs| outs.iter().find(|o| o["scriptPubKey"]["hex"].as_str() == Some(spk_hex)))
        .unwrap_or_else(|| fail("funding output not found"));
    (o["n"].as_u64().unwrap() as u32, (o["value"].as_f64().unwrap() * 1e8).round() as u64)
}

#[allow(clippy::too_many_arguments)]
fn build_spend(
    node: &BitcoinRPC,
    wallet: &BitcoinRPC,
    secp: &Secp256k1<bitcoin::secp256k1::All>,
    sk: &SecretKey,
    witness_script: &ScriptBuf,
    txid: &str,
    vout: u32,
    value_sat: u64,
    version: Version,
    sequence: Sequence,
    lock_time: LockTime,
) -> String {
    let dest = rpc(wallet, "getnewaddress", vec![json!("timelock_return"), json!("bech32m")]);
    let dest_spk = Address::from_str(dest.as_str().unwrap())
        .unwrap()
        .require_network(Network::Regtest)
        .unwrap()
        .script_pubkey();
    let out_sat = value_sat.checked_sub(FEE_SAT).unwrap_or(0);
    if out_sat == 0 {
        fail("funded amount too small for fee");
    }
    let mut tx = Transaction {
        version,
        lock_time,
        input: vec![TxIn {
            previous_output: OutPoint {
                txid: Txid::from_str(txid).unwrap(),
                vout,
            },
            script_sig: ScriptBuf::new(),
            sequence,
            witness: Witness::new(),
        }],
        output: vec![TxOut {
            value: Amount::from_sat(out_sat),
            script_pubkey: dest_spk,
        }],
    };
    let sighash = SighashCache::new(&tx)
        .p2wsh_signature_hash(0, witness_script, Amount::from_sat(value_sat), EcdsaSighashType::All)
        .unwrap_or_else(|e| fail(&e.to_string()));
    let msg = Message::from_digest(sighash.to_byte_array());
    let mut sig = secp.sign_ecdsa_low_r(&msg, sk).serialize_der().to_vec();
    sig.push(EcdsaSighashType::All as u8);
    let mut w = Witness::new();
    w.push(&sig);
    w.push(witness_script.as_bytes());
    tx.input[0].witness = w;
    let _ = node;
    serialize_hex(&tx)
}

fn main() {
    println!("=== Lab 05: Timelocks (Rust) ===");
    let secp = Secp256k1::new();

    step!("[Step 1] Bootstrapping lab wallet & funds ... ");
    let res = match bootstrap_lab(None) {
        Ok(r) => r,
        Err(e) => fail(&e),
    };
    let (node, wallet) = (res.rpc, res.wallet_rpc);
    let sk = SecretKey::from_str(SK_A_HEX).unwrap();
    let pk = PublicKey::new(sk.public_key(&secp));
    let miner = rpc(&wallet, "getnewaddress", vec![json!("lab05_miner"), json!("bech32m")])
        .as_str()
        .unwrap()
        .to_string();
    println!("✓");

    // ---------- CLTV ----------
    step!("[Step 2] CLTV: building <height> OP_CLTV script ... ");
    let lock_height = rpc(&node, "getblockcount", vec![]).as_u64().unwrap() + 10;
    let cltv_script = Builder::new()
        .push_int(lock_height as i64)
        .push_opcode(OP_CLTV)
        .push_opcode(OP_DROP)
        .push_key(&pk)
        .push_opcode(OP_CHECKSIG)
        .into_script();
    let cltv_addr = Address::p2wsh(&cltv_script, Network::Regtest);
    println!("✓ (lock at height {})", lock_height);

    step!("[Step 3] CLTV: deriving P2WSH address & validating with node ... ");
    if rpc(&node, "validateaddress", vec![json!(cltv_addr.to_string())])["isvalid"].as_bool() != Some(true) {
        fail("node rejected CLTV address");
    }
    println!("✓ ({}…)", &cltv_addr.to_string()[..14]);

    step!("[Step 4] CLTV: funding ({} BTC) & mining 1 block ... ", FUND_BTC);
    let cltv_txid = rpc(&wallet, "sendtoaddress", vec![json!(cltv_addr.to_string()), json!(FUND_BTC)])
        .as_str()
        .unwrap()
        .to_string();
    rpc(&wallet, "generatetoaddress", vec![json!(1), json!(miner)]);
    let cltv_spk = cltv_addr.script_pubkey().to_hex_string();
    let (cltv_vout, cltv_sat) = find_vout(&node, &cltv_txid, &cltv_spk);
    println!("✓ (vout {})", cltv_vout);

    step!("[Step 5] CLTV: early spend rejected by testmempoolaccept ... ");
    let raw = build_spend(&node, &wallet, &secp, &sk, &cltv_script, &cltv_txid, cltv_vout, cltv_sat,
        Version::TWO, Sequence::from_consensus(0xFFFF_FFFE),
        LockTime::from_height(lock_height as u32).unwrap());
    let res_early = rpc(&node, "testmempoolaccept", vec![json!([raw])]);
    if res_early[0]["allowed"].as_bool() != Some(false) {
        fail(&format!("expected rejection, got {}", res_early));
    }
    println!("✓ (allowed: false — {})", res_early[0]["reject-reason"].as_str().unwrap_or("?"));

    step!("[Step 6] CLTV: mining to the lock height ... ");
    let cur = rpc(&node, "getblockcount", vec![]).as_u64().unwrap();
    if lock_height + 1 > cur {
        rpc(&wallet, "generatetoaddress", vec![json!(lock_height + 1 - cur), json!(miner)]);
    }
    println!("✓ (height {})", rpc(&node, "getblockcount", vec![]).as_u64().unwrap());

    step!("[Step 7] CLTV: spend accepted, broadcast & confirmed ... ");
    let raw = build_spend(&node, &wallet, &secp, &sk, &cltv_script, &cltv_txid, cltv_vout, cltv_sat,
        Version::TWO, Sequence::from_consensus(0xFFFF_FFFE),
        LockTime::from_height(lock_height as u32).unwrap());
    if rpc(&node, "testmempoolaccept", vec![json!([raw])])[0]["allowed"].as_bool() != Some(true) {
        fail("mempool rejected the mature CLTV spend");
    }
    let sp = rpc(&node, "sendrawtransaction", vec![json!(raw)]).as_str().unwrap().to_string();
    rpc(&wallet, "generatetoaddress", vec![json!(1), json!(miner)]);
    let confs = rpc(&node, "getrawtransaction", vec![json!(sp), json!(true)])["confirmations"].as_u64().unwrap_or(0);
    if confs < 1 {
        fail("CLTV spend not confirmed");
    }
    println!("✓ (confirmations: {})", confs);

    // ---------- CSV ----------
    step!("[Step 8] CSV: building <3> OP_CSV script ... ");
    let csv_script = Builder::new()
        .push_int(CSV_DELAY)
        .push_opcode(OP_CSV)
        .push_opcode(OP_DROP)
        .push_key(&pk)
        .push_opcode(OP_CHECKSIG)
        .into_script();
    if csv_script.to_hex_string() != EXPECT_CSV_SCRIPT {
        fail(&format!("CSV script {} != canonical", csv_script.to_hex_string()));
    }
    let csv_addr = Address::p2wsh(&csv_script, Network::Regtest);
    println!("✓ (matches canonical)");

    step!("[Step 9] CSV: deriving P2WSH address & validating with node ... ");
    if csv_addr.to_string() != EXPECT_CSV_ADDR {
        fail(&format!("{} != {}", csv_addr, EXPECT_CSV_ADDR));
    }
    if rpc(&node, "validateaddress", vec![json!(csv_addr.to_string())])["isvalid"].as_bool() != Some(true) {
        fail("node rejected CSV address");
    }
    println!("✓ ({}…)", &csv_addr.to_string()[..14]);

    step!("[Step 10] CSV: funding ({} BTC) & mining 1 block ... ", FUND_BTC);
    let csv_txid = rpc(&wallet, "sendtoaddress", vec![json!(csv_addr.to_string()), json!(FUND_BTC)])
        .as_str()
        .unwrap()
        .to_string();
    rpc(&wallet, "generatetoaddress", vec![json!(1), json!(miner)]);
    let csv_spk = csv_addr.script_pubkey().to_hex_string();
    let (csv_vout, csv_sat) = find_vout(&node, &csv_txid, &csv_spk);
    println!("✓ (vout {}, 1 confirmation)", csv_vout);

    step!("[Step 11] CSV: early spend rejected by testmempoolaccept ... ");
    let raw = build_spend(&node, &wallet, &secp, &sk, &csv_script, &csv_txid, csv_vout, csv_sat,
        Version::TWO, Sequence::from_consensus(CSV_DELAY as u32), LockTime::ZERO);
    let res_early = rpc(&node, "testmempoolaccept", vec![json!([raw])]);
    if res_early[0]["allowed"].as_bool() != Some(false) {
        fail(&format!("expected rejection, got {}", res_early));
    }
    println!("✓ (allowed: false — {})", res_early[0]["reject-reason"].as_str().unwrap_or("?"));

    step!("[Step 12] CSV: mining 3 blocks to satisfy the relative delay ... ");
    rpc(&wallet, "generatetoaddress", vec![json!(CSV_DELAY), json!(miner)]);
    println!("✓");

    step!("[Step 13] CSV: spend accepted, broadcast & confirmed ... ");
    let raw = build_spend(&node, &wallet, &secp, &sk, &csv_script, &csv_txid, csv_vout, csv_sat,
        Version::TWO, Sequence::from_consensus(CSV_DELAY as u32), LockTime::ZERO);
    if rpc(&node, "testmempoolaccept", vec![json!([raw])])[0]["allowed"].as_bool() != Some(true) {
        fail("mempool rejected the mature CSV spend");
    }
    let sp = rpc(&node, "sendrawtransaction", vec![json!(raw)]).as_str().unwrap().to_string();
    rpc(&wallet, "generatetoaddress", vec![json!(1), json!(miner)]);
    let confs = rpc(&node, "getrawtransaction", vec![json!(sp), json!(true)])["confirmations"].as_u64().unwrap_or(0);
    if confs < 1 {
        fail("CSV spend not confirmed");
    }
    println!("✓ (confirmations: {})", confs);

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}
