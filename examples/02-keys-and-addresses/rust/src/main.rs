use bitcoin::bip32::{DerivationPath, Xpriv};
use bitcoin::secp256k1::Secp256k1;
use bitcoin::{Address, Network};
use bitcoin_sandbox_common::bootstrap_lab;
use serde_json::json;
use std::io::{self, Write};
use std::process;
use std::str::FromStr;

fn main() {
    println!("=== Lab 02: Keys & Addresses (Rust) ===");

    // Step 1: Bootstrap lab
    print!("[Step 1] Bootstrapping lab wallet & RPC connection ... ");
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
    let secp = Secp256k1::new();

    // Step 2: Deterministic test seed (derived from 12 words: abandon abandon ... about)
    print!("[Step 2] Generating BIP39 binary seed ... ");
    io::stdout().flush().unwrap();
    // 64-byte binary seed for 'abandon ... about'
    let seed_hex = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4";
    let seed = hex::decode(seed_hex).unwrap_or_else(|_| vec![0u8; 64]);
    println!("✓ (64 bytes)");

    // Step 3: BIP32 Root Key
    print!("[Step 3] Deriving BIP32 root key ... ");
    io::stdout().flush().unwrap();
    let root = match Xpriv::new_master(Network::Regtest, &seed) {
        Ok(k) => k,
        Err(e) => {
            eprintln!("\n✗ FAILURE: {}", e);
            process::exit(1);
        }
    };
    println!("✓ (fingerprint: {})", root.fingerprint(&secp));

    // Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
    print!("[Step 4] Deriving BIP44 Legacy P2PKH address ... ");
    io::stdout().flush().unwrap();
    let path_p2pkh = DerivationPath::from_str("m/44'/1'/0'/0/0").unwrap();
    let child_p2pkh = root.derive_priv(&secp, &path_p2pkh).unwrap();
    let pub_p2pkh = child_p2pkh.to_priv().public_key(&secp);
    let addr_p2pkh = Address::p2pkh(pub_p2pkh, Network::Regtest);
    println!("✓ ({})", addr_p2pkh);

    // Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
    print!("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ");
    io::stdout().flush().unwrap();
    let path_p2wpkh = DerivationPath::from_str("m/84'/1'/0'/0/0").unwrap();
    let child_p2wpkh = root.derive_priv(&secp, &path_p2wpkh).unwrap();
    let pub_p2wpkh = child_p2wpkh.to_priv().public_key(&secp);
    let addr_p2wpkh = Address::p2wpkh(&pub_p2wpkh, Network::Regtest);
    println!("✓ ({})", addr_p2wpkh);

    // Step 6: Taproot P2TR (m/86'/1'/0'/0/0)
    print!("[Step 6] Deriving BIP86 Taproot P2TR address ... ");
    io::stdout().flush().unwrap();
    let path_p2tr = DerivationPath::from_str("m/86'/1'/0'/0/0").unwrap();
    let child_p2tr = root.derive_priv(&secp, &path_p2tr).unwrap();
    let keypair = child_p2tr.to_priv().keypair(&secp);
    let (x_only, _parity) = keypair.x_only_public_key();
    let addr_p2tr = Address::p2tr(&secp, x_only, None, Network::Regtest);
    println!("✓ ({})", addr_p2tr);

    // Step 7: Validate derived addresses with Bitcoin Core node
    print!("[Step 7] Validating addresses against Bitcoin Core node ... ");
    io::stdout().flush().unwrap();
    for addr in &[addr_p2pkh.to_string(), addr_p2wpkh.to_string(), addr_p2tr.to_string()] {
        let val = match node_rpc.call("validateaddress", vec![json!(addr)]) {
            Ok(v) => v,
            Err(e) => {
                eprintln!("\n✗ FAILURE: {}", e);
                process::exit(1);
            }
        };
        let is_valid = val.get("isvalid").and_then(|v| v.as_bool()).unwrap_or(false);
        if !is_valid {
            eprintln!("\n✗ FAILURE: address {} reported invalid by node", addr);
            process::exit(1);
        }
    }
    println!("✓ (all addresses valid)");

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}

// Minimal hex decoder helper to avoid external hex crate dependency
mod hex {
    pub fn decode(s: &str) -> Result<Vec<u8>, ()> {
        if s.len() % 2 != 0 {
            return Err(());
        }
        (0..s.len())
            .step_by(2)
            .map(|i| u8::from_str_radix(&s[i..i + 2], 16).map_err(|_| ()))
            .collect()
    }
}
