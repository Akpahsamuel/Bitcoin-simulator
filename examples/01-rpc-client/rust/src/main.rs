use bitcoin_sandbox_common::bootstrap_lab;
use serde_json::json;
use std::io::{self, Write};
use std::process;

fn main() {
    println!("=== Lab 01: RPC Client (Rust) ===");

    // Step 1: Bootstrap lab wallet & coinbase maturity
    print!("[Step 1] Bootstrapping lab wallet and initial funds ... ");
    io::stdout().flush().unwrap();
    let bootstrap_res = match bootstrap_lab(None) {
        Ok(res) => {
            println!("✓");
            res
        }
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            println!("======================================================");
            println!("Result: FAIL ✗");
            println!("======================================================");
            process::exit(1);
        }
    };

    let node_rpc = bootstrap_res.rpc;
    let wallet_rpc = bootstrap_res.wallet_rpc;

    // Step 2: Query blockchain info via JSON-RPC
    print!("[Step 2] Querying getblockchaininfo via JSON-RPC ... ");
    io::stdout().flush().unwrap();
    let chain_info = match node_rpc.call("getblockchaininfo", vec![]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };

    let chain = chain_info.get("chain").and_then(|c| c.as_str()).unwrap_or("");
    let blocks = chain_info.get("blocks").and_then(|b| b.as_u64()).unwrap_or(0);
    if chain != "regtest" || blocks < 101 {
        eprintln!("\n✗ FAILURE: unexpected chain {} or blocks {}", chain, blocks);
        process::exit(1);
    }
    println!("✓ (chain: {}, blocks: {})", chain, blocks);

    // Step 3: Generate fresh address and mine 1 block
    print!("[Step 3] Generating fresh address and mining 1 block ... ");
    io::stdout().flush().unwrap();
    let fresh_addr = match wallet_rpc.call("getnewaddress", vec![json!("lab01_test"), json!("bech32m")]) {
        Ok(v) => v.as_str().unwrap_or("").to_string(),
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };

    let block_hashes = match wallet_rpc.call("generatetoaddress", vec![json!(1), json!(fresh_addr)]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let hash_str = block_hashes
        .as_array()
        .and_then(|arr| arr.first())
        .and_then(|h| h.as_str())
        .unwrap_or("");

    let new_blocks_val = node_rpc.call("getblockcount", vec![]).unwrap_or(json!(0));
    let new_blocks = new_blocks_val.as_u64().unwrap_or(0);
    println!("✓ (new height: {}, mined block: {}...)", new_blocks, &hash_str[..16.min(hash_str.len())]);

    // Step 4: Query wallet balance
    print!("[Step 4] Querying wallet balance via getbalance ... ");
    io::stdout().flush().unwrap();
    let balance_val = match wallet_rpc.call("getbalance", vec![]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let balance = balance_val.as_f64().unwrap_or(0.0);
    if balance <= 0.0 {
        eprintln!("\n✗ FAILURE: expected spendable balance > 0, got {}", balance);
        process::exit(1);
    }
    println!("✓ (balance: {} BTC)", balance);

    // Step 5: Query unauthenticated REST API
    print!("[Step 5] Querying unauthenticated REST API (/rest/chaininfo.json) ... ");
    io::stdout().flush().unwrap();
    let rest_info = match node_rpc.get_rest("chaininfo.json") {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let rest_chain = rest_info.get("chain").and_then(|c| c.as_str()).unwrap_or("");
    let rest_blocks = rest_info.get("blocks").and_then(|b| b.as_u64()).unwrap_or(0);
    if rest_chain != "regtest" || rest_blocks != new_blocks {
        eprintln!("\n✗ FAILURE: REST info verification failed");
        process::exit(1);
    }
    println!("✓");

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}
