//! Lab 06: ZMQ Listener (Rust Reference Implementation)
//! Subscribe to Bitcoin Core's rawtx / rawblock ZeroMQ feeds, trigger one
//! transaction and one block, and verify the pushed bytes against the node.

use bitcoin_sandbox_common::{bootstrap_lab, get_config, BitcoinRPC};
use serde_json::{json, Value};
use std::io::{self, Write};
use std::process;
use std::thread::sleep;
use std::time::Duration;

const RECV_TIMEOUT_MS: i32 = 15_000;

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

fn subscribe(ctx: &zmq::Context, endpoint: &str, topic: &str) -> zmq::Socket {
    let sock = ctx.socket(zmq::SUB).unwrap_or_else(|e| fail(&e.to_string()));
    sock.set_rcvtimeo(RECV_TIMEOUT_MS).ok();
    sock.set_subscribe(topic.as_bytes()).ok();
    sock.connect(endpoint).unwrap_or_else(|e| fail(&format!("connect {}: {}", endpoint, e)));
    sock
}

fn recv3(sock: &zmq::Socket, what: &str) -> (Vec<u8>, Vec<u8>, Vec<u8>) {
    match sock.recv_multipart(0) {
        Ok(parts) if parts.len() >= 3 => (parts[0].clone(), parts[1].clone(), parts[2].clone()),
        Ok(_) => fail(&format!("{}: unexpected frame count", what)),
        Err(_) => fail(&format!(
            "no {} notification within {} ms — is the matching zmqpub* set? (see docs/troubleshooting.md)",
            what, RECV_TIMEOUT_MS
        )),
    }
}

fn seq(bytes: &[u8]) -> u32 {
    let mut b = [0u8; 4];
    b.copy_from_slice(&bytes[..4]);
    u32::from_le_bytes(b)
}

fn main() {
    println!("=== Lab 06: ZMQ Listener (Rust) ===");

    step!("[Step 1] Bootstrapping lab wallet & funds ... ");
    let res = match bootstrap_lab(None) {
        Ok(r) => r,
        Err(e) => fail(&e),
    };
    let (node, wallet) = (res.rpc, res.wallet_rpc);
    let cfg = get_config();
    println!("✓");

    step!(
        "[Step 2] Subscribing to rawtx ({}) and rawblock ({}) ... ",
        cfg.zmq_rawtx, cfg.zmq_rawblock
    );
    let ctx = zmq::Context::new();
    let sub_tx = subscribe(&ctx, &cfg.zmq_rawtx, "rawtx");
    let sub_block = subscribe(&ctx, &cfg.zmq_rawblock, "rawblock");
    sleep(Duration::from_millis(500)); // let the SUB subscriptions propagate
    println!("✓");

    step!("[Step 3] Broadcasting a transaction ... ");
    let dest = rpc(&wallet, "getnewaddress", vec![json!("zmq_probe"), json!("bech32m")])
        .as_str()
        .unwrap()
        .to_string();
    let sent_txid = rpc(&wallet, "sendtoaddress", vec![json!(dest), json!(0.01)])
        .as_str()
        .unwrap()
        .to_string();
    println!("✓ (txid: {}…)", &sent_txid[..16]);

    step!("[Step 4] Received rawtx frame & verified txid ... ");
    let (topic, body, s) = recv3(&sub_tx, "rawtx");
    if topic != b"rawtx" {
        fail("unexpected topic on rawtx socket");
    }
    let decoded = rpc(&node, "decoderawtransaction", vec![json!(hex::encode(&body))]);
    if decoded["txid"].as_str() != Some(sent_txid.as_str()) {
        fail(&format!("rawtx txid {} != {}", decoded["txid"], sent_txid));
    }
    println!("✓ (seq {})", seq(&s));

    step!("[Step 5] Mining 1 block ... ");
    let miner = rpc(&wallet, "getnewaddress", vec![json!("lab06_miner"), json!("bech32m")])
        .as_str()
        .unwrap()
        .to_string();
    let block_hash = rpc(&wallet, "generatetoaddress", vec![json!(1), json!(miner)])[0]
        .as_str()
        .unwrap()
        .to_string();
    println!("✓ (hash: {}…)", &block_hash[..16]);

    step!("[Step 6] Received rawblock frame & verified against getblock ... ");
    let (topic, body, s) = recv3(&sub_block, "rawblock");
    if topic != b"rawblock" {
        fail("unexpected topic on rawblock socket");
    }
    let expected = rpc(&node, "getblock", vec![json!(block_hash), json!(0)]);
    if expected.as_str() != Some(hex::encode(&body).as_str()) {
        fail("rawblock payload does not match getblock");
    }
    println!("✓ (seq {})", seq(&s));

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}

/// Tiny hex encoder — avoids an extra crate for two round-trips.
mod hex {
    pub fn encode(bytes: &[u8]) -> String {
        let mut s = String::with_capacity(bytes.len() * 2);
        for b in bytes {
            s.push_str(&format!("{:02x}", b));
        }
        s
    }
}
