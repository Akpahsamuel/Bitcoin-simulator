//! Lab 04: Multisig Escrow (Rust Reference Implementation)
//! 2-of-3 P2WSH built from raw script, funded from the lab wallet, then spent by
//! assembling the witness stack [<empty>, sigA, sigB, witnessScript] by hand.

use bitcoin::absolute::LockTime;
use bitcoin::consensus::encode::serialize_hex;
use bitcoin::hashes::Hash;
use bitcoin::opcodes::all::{OP_CHECKMULTISIG, OP_PUSHNUM_2, OP_PUSHNUM_3};
use bitcoin::script::Builder;
use bitcoin::secp256k1::{Message, Secp256k1, SecretKey};
use bitcoin::sighash::{EcdsaSighashType, SighashCache};
use bitcoin::transaction::Version;
use bitcoin::{
    Address, Amount, Network, OutPoint, PublicKey, ScriptBuf, Sequence, Transaction, TxIn, TxOut,
    Txid, Witness,
};
use bitcoin_sandbox_common::bootstrap_lab;
use serde_json::json;
use std::io::{self, Write};
use std::process;
use std::str::FromStr;

const SK_HEX: [&str; 3] = [
    "1111111111111111111111111111111111111111111111111111111111111111",
    "2222222222222222222222222222222222222222222222222222222222222222",
    "3333333333333333333333333333333333333333333333333333333333333333",
];
const EXPECT_WITNESS_SCRIPT: &str = "5221034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa2102466d7fcae563e5cb09a0d1870bb580344804617879a14949cf22285f1bae3f2721023c72addb4fdf09af94f0c94d7fe92a386a7e70cf8a1d85916386bb2535c7b1b153ae";
const EXPECT_P2WSH: &str = "bcrt1qpy8yjjs2l5neewx722mxve9w6m77zqsu7rldukggseflhwralerqh6ma0d";

const FUND_BTC: f64 = 0.5;
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

fn main() {
    println!("=== Lab 04: Multisig Escrow (Rust) ===");
    let secp = Secp256k1::new();

    // Step 1
    step!("[Step 1] Bootstrapping lab wallet & funds ... ");
    let res = match bootstrap_lab(None) {
        Ok(r) => r,
        Err(e) => fail(&e),
    };
    let (node_rpc, wallet_rpc) = (res.rpc, res.wallet_rpc);
    println!("✓");

    // Step 2
    step!("[Step 2] Deriving 3 escrow keypairs from test vectors ... ");
    let sks: Vec<SecretKey> = SK_HEX
        .iter()
        .map(|h| SecretKey::from_str(h).unwrap_or_else(|_| fail("bad test secret")))
        .collect();
    let pubs: Vec<PublicKey> = sks
        .iter()
        .map(|sk| PublicKey::new(sk.public_key(&secp)))
        .collect();
    println!(
        "✓ ({})",
        pubs.iter()
            .map(|p| format!("{}…", &p.to_string()[..10]))
            .collect::<Vec<_>>()
            .join(", ")
    );

    // Step 3
    step!("[Step 3] Building 2-of-3 witness script ... ");
    let witness_script: ScriptBuf = Builder::new()
        .push_opcode(OP_PUSHNUM_2)
        .push_key(&pubs[0])
        .push_key(&pubs[1])
        .push_key(&pubs[2])
        .push_opcode(OP_PUSHNUM_3)
        .push_opcode(OP_CHECKMULTISIG)
        .into_script();
    if witness_script.to_hex_string() != EXPECT_WITNESS_SCRIPT {
        fail(&format!("witness script {} != canonical", witness_script.to_hex_string()));
    }
    println!("✓ (matches canonical)");

    // Step 4
    step!("[Step 4] Deriving P2WSH address & validating with node ... ");
    let addr = Address::p2wsh(&witness_script, Network::Regtest);
    if addr.to_string() != EXPECT_P2WSH {
        fail(&format!("address {} != {}", addr, EXPECT_P2WSH));
    }
    let val = node_rpc
        .call("validateaddress", vec![json!(addr.to_string())])
        .unwrap_or_else(|e| fail(&e));
    if val.get("isvalid").and_then(|v| v.as_bool()) != Some(true) {
        fail("node rejected the P2WSH address");
    }
    println!("✓ ({}…)", &addr.to_string()[..14]);

    // Step 5
    step!("[Step 5] Funding the multisig ({} BTC) & mining 1 block ... ", FUND_BTC);
    let fund_txid = wallet_rpc
        .call("sendtoaddress", vec![json!(addr.to_string()), json!(FUND_BTC)])
        .unwrap_or_else(|e| fail(&e));
    let fund_txid = fund_txid.as_str().unwrap_or_else(|| fail("no txid")).to_string();
    let miner = wallet_rpc
        .call("getnewaddress", vec![json!("lab04_miner"), json!("bech32m")])
        .unwrap_or_else(|e| fail(&e));
    let miner = miner.as_str().unwrap().to_string();
    wallet_rpc
        .call("generatetoaddress", vec![json!(1), json!(miner)])
        .unwrap_or_else(|e| fail(&e));
    println!("✓ (txid: {}…)", &fund_txid[..16]);

    // Step 6
    step!("[Step 6] Locating the funding UTXO by scriptPubKey ... ");
    let spk_hex = addr.script_pubkey().to_hex_string();
    let fund_tx = node_rpc
        .call("getrawtransaction", vec![json!(fund_txid), json!(true)])
        .unwrap_or_else(|e| fail(&e));
    let vout = fund_tx["vout"]
        .as_array()
        .and_then(|outs| {
            outs.iter().find(|o| {
                o["scriptPubKey"]["hex"].as_str() == Some(spk_hex.as_str())
            })
        })
        .unwrap_or_else(|| fail("funding output not found"));
    let fund_vout = vout["n"].as_u64().unwrap() as u32;
    let fund_btc = vout["value"].as_f64().unwrap();
    let fund_sat = (fund_btc * 1e8).round() as u64;
    println!("✓ (vout: {}, {} BTC)", fund_vout, fund_btc);

    // Step 7
    step!("[Step 7] Building the spend transaction ... ");
    let return_addr = node_rpc
        .call("getnewaddress", vec![json!("escrow_return"), json!("bech32m")])
        .unwrap_or_else(|_| {
            wallet_rpc
                .call("getnewaddress", vec![json!("escrow_return"), json!("bech32m")])
                .unwrap_or_else(|e| fail(&e))
        });
    let return_addr = return_addr.as_str().unwrap().to_string();
    let return_spk = Address::from_str(&return_addr)
        .unwrap_or_else(|_| fail("bad return address"))
        .require_network(Network::Regtest)
        .unwrap_or_else(|_| fail("return address wrong network"))
        .script_pubkey();
    let out_sat = fund_sat.checked_sub(FEE_SAT).unwrap_or(0);
    if out_sat == 0 {
        fail("funded amount too small for fee");
    }
    let mut tx = Transaction {
        version: Version::TWO,
        lock_time: LockTime::ZERO,
        input: vec![TxIn {
            previous_output: OutPoint {
                txid: Txid::from_str(&fund_txid).unwrap_or_else(|_| fail("bad txid")),
                vout: fund_vout,
            },
            script_sig: ScriptBuf::new(),
            sequence: Sequence::MAX,
            witness: Witness::new(),
        }],
        output: vec![TxOut {
            value: Amount::from_sat(out_sat),
            script_pubkey: return_spk,
        }],
    };
    println!("✓");

    // Step 8
    step!("[Step 8] BIP143 sighash + signing with keys A and B ... ");
    let sighash = SighashCache::new(&tx)
        .p2wsh_signature_hash(0, &witness_script, Amount::from_sat(fund_sat), EcdsaSighashType::All)
        .unwrap_or_else(|e| fail(&e.to_string()));
    let msg = Message::from_digest(sighash.to_byte_array());
    let sign = |sk: &SecretKey| -> Vec<u8> {
        let mut v = secp.sign_ecdsa_low_r(&msg, sk).serialize_der().to_vec();
        v.push(EcdsaSighashType::All as u8);
        v
    };
    let sig_a = sign(&sks[0]);
    let sig_b = sign(&sks[1]);
    println!("✓ (sigA {}B, sigB {}B)", sig_a.len(), sig_b.len());

    // Step 9
    step!("[Step 9] Assembling witness [<empty>, sigA, sigB, script] ... ");
    let mut w = Witness::new();
    w.push([] as [u8; 0]);
    w.push(&sig_a);
    w.push(&sig_b);
    w.push(witness_script.as_bytes());
    tx.input[0].witness = w;
    let raw_hex = serialize_hex(&tx);
    println!("✓ (tx {} bytes)", raw_hex.len() / 2);

    // Step 10
    step!("[Step 10] testmempoolaccept ... ");
    let accept = node_rpc
        .call("testmempoolaccept", vec![json!([raw_hex])])
        .unwrap_or_else(|e| fail(&e));
    if accept[0]["allowed"].as_bool() != Some(true) {
        fail(&format!("mempool rejected: {}", accept));
    }
    println!("✓ (allowed: true)");

    // Step 11
    step!("[Step 11] Broadcasting & mining 1 block ... ");
    let spend_txid = node_rpc
        .call("sendrawtransaction", vec![json!(raw_hex)])
        .unwrap_or_else(|e| fail(&e));
    let spend_txid = spend_txid.as_str().unwrap().to_string();
    wallet_rpc
        .call("generatetoaddress", vec![json!(1), json!(miner)])
        .unwrap_or_else(|e| fail(&e));
    let details = node_rpc
        .call("getrawtransaction", vec![json!(spend_txid), json!(true)])
        .unwrap_or_else(|e| fail(&e));
    let confs = details["confirmations"].as_u64().unwrap_or(0);
    if confs < 1 {
        fail(&format!("expected confirmations >= 1, got {}", confs));
    }
    println!("✓ (confirmations: {})", confs);

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}
