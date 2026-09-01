use bitcoin_sandbox_common::bootstrap_lab;
use serde_json::json;
use std::io::{self, Write};
use std::process;

fn main() {
    println!("=== Lab 03: Raw Transactions (Rust) ===");

    // Step 1: Bootstrap lab
    print!("[Step 1] Bootstrapping lab wallet & ensure spendable UTXOs ... ");
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

    // Step 2: Select UTXO from listunspent
    print!("[Step 2] Selecting UTXO from listunspent ... ");
    io::stdout().flush().unwrap();
    let unspent_val = match wallet_rpc.call("listunspent", vec![json!(1), json!(9999999)]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let unspent = unspent_val.as_array().filter(|a| !a.is_empty()).unwrap_or_else(|| {
        eprintln!("\n✗ FAILURE: No spendable UTXOs found in wallet");
        process::exit(1);
    });

    let utxo = &unspent[0];
    let utxo_txid = utxo.get("txid").and_then(|t| t.as_str()).unwrap_or("");
    let utxo_vout = utxo.get("vout").and_then(|v| v.as_u64()).unwrap_or(0);
    let utxo_amount = utxo.get("amount").and_then(|a| a.as_f64()).unwrap_or(0.0);
    println!("✓ (txid: {}..., vout: {}, amount: {} BTC)", &utxo_txid[..16.min(utxo_txid.len())], utxo_vout, utxo_amount);

    // Step 3: Construct raw transaction
    print!("[Step 3] Constructing raw transaction hex (createrawtransaction) ... ");
    io::stdout().flush().unwrap();
    let recipient_addr = match wallet_rpc.call("getnewaddress", vec![json!("recipient"), json!("bech32m")]) {
        Ok(v) => v.as_str().unwrap_or("").to_string(),
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let change_addr = match wallet_rpc.call("getnewaddress", vec![json!("change"), json!("bech32m")]) {
        Ok(v) => v.as_str().unwrap_or("").to_string(),
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };

    let send_amount = 1.5;
    let fee = 0.0001;
    let change_amount = ((utxo_amount - send_amount - fee) * 1e8).round() / 1e8;
    if change_amount <= 0.0 {
        eprintln!("\n✗ FAILURE: UTXO amount insufficient for send + fee");
        process::exit(1);
    }

    let inputs = json!([{"txid": utxo_txid, "vout": utxo_vout}]);
    let outputs = json!([
        { recipient_addr: send_amount },
        { change_addr: change_amount }
    ]);

    let raw_tx_hex_val = match wallet_rpc.call("createrawtransaction", vec![inputs, outputs]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let raw_tx_hex = raw_tx_hex_val.as_str().unwrap_or("");
    println!("✓ (hex length: {})", raw_tx_hex.len());

    // Step 4: Sign transaction inputs
    print!("[Step 4] Signing transaction inputs (signrawtransactionwithwallet) ... ");
    io::stdout().flush().unwrap();
    let sign_val = match wallet_rpc.call("signrawtransactionwithwallet", vec![json!(raw_tx_hex)]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let complete = sign_val.get("complete").and_then(|c| c.as_bool()).unwrap_or(false);
    let signed_hex = sign_val.get("hex").and_then(|h| h.as_str()).unwrap_or("");
    if !complete || signed_hex.is_empty() {
        eprintln!("\n✗ FAILURE: transaction signing was incomplete");
        process::exit(1);
    }
    println!("✓ (complete: true)");

    // Step 5: Verify transaction with testmempoolaccept
    print!("[Step 5] Verifying transaction with testmempoolaccept ... ");
    io::stdout().flush().unwrap();
    let mempool_val = match node_rpc.call("testmempoolaccept", vec![json!([signed_hex])]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let allowed = mempool_val
        .as_array()
        .and_then(|arr| arr.first())
        .and_then(|entry| entry.get("allowed"))
        .and_then(|a| a.as_bool())
        .unwrap_or(false);
    if !allowed {
        eprintln!("\n✗ FAILURE: mempool rejected transaction: {:?}", mempool_val);
        process::exit(1);
    }
    println!("✓ (allowed: true)");

    // Step 6: Broadcast transaction via sendrawtransaction
    print!("[Step 6] Broadcasting transaction via sendrawtransaction ... ");
    io::stdout().flush().unwrap();
    let txid_val = match node_rpc.call("sendrawtransaction", vec![json!(signed_hex)]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let txid = txid_val.as_str().unwrap_or("");
    if txid.len() != 64 {
        eprintln!("\n✗ FAILURE: invalid broadcast txid: {}", txid);
        process::exit(1);
    }
    println!("✓ (txid: {})", txid);

    // Step 7: Mine 1 block & verify confirmation
    print!("[Step 7] Mining 1 block & verifying confirmation ... ");
    io::stdout().flush().unwrap();
    let miner_addr = match wallet_rpc.call("getnewaddress", vec![json!("miner"), json!("bech32m")]) {
        Ok(v) => v.as_str().unwrap_or("").to_string(),
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let _ = wallet_rpc.call("generatetoaddress", vec![json!(1), json!(miner_addr)]);

    let tx_details = match wallet_rpc.call("getrawtransaction", vec![json!(txid), json!(true)]) {
        Ok(v) => v,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    let confs = tx_details.get("confirmations").and_then(|c| c.as_u64()).unwrap_or(0);
    if confs < 1 {
        eprintln!("\n✗ FAILURE: expected confirmations >= 1, got {}", confs);
        process::exit(1);
    }
    println!("✓ (confirmations: {})", confs);

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}
