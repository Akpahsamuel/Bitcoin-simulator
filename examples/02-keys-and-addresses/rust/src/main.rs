use bitcoin::bip32::{DerivationPath, Xpriv};
use bitcoin::key::UntweakedPublicKey;
use bitcoin::secp256k1::Secp256k1;
use bitcoin::{Address, CompressedPublicKey, Network, PublicKey};
use bitcoin_sandbox_common::bootstrap_lab;
use serde_json::json;
use std::io::{self, Write};
use std::process;
use std::str::FromStr;

// Deterministic test vector: the canonical "abandon ... about" mnemonic on regtest.
// Expected values are cross-checked against embit (Python), bitcoinjs-lib (TypeScript)
// and btcd (Go) — every port must derive exactly these.
const SEED_HEX: &str = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4";
const EXPECT_FINGERPRINT: &str = "73c5da0a";
const EXPECT_P2PKH: &str = "mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV";
const EXPECT_P2WPKH: &str = "bcrt1q6rz28mcfaxtmd6v789l9rrlrusdprr9pz3cppk";
const EXPECT_P2TR: &str = "bcrt1p8wpt9v4frpf3tkn0srd97pksgsxc5hs52lafxwru9kgeephvs7rqjeprhg";

fn fail(msg: &str) -> ! {
    eprintln!("\n✗ FAILURE: {}", msg);
    println!("======================================================");
    println!("Result: FAIL ✗");
    println!("======================================================");
    process::exit(1);
}

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
        Err(e) => fail(&e),
    };

    let node_rpc = bootstrap_res.rpc;
    let secp = Secp256k1::new();

    // Step 2: BIP39 mnemonic -> 512-bit root seed (test vector seed shown directly).
    print!("[Step 2] Generating BIP39 mnemonic & root seed ... ");
    io::stdout().flush().unwrap();
    let seed = hex_decode(SEED_HEX).unwrap_or_else(|_| fail("bad seed hex"));
    if seed.len() != 64 {
        fail("expected 64-byte seed");
    }
    println!("✓ (12 words, {}-byte seed)", seed.len());

    // Step 3: BIP32 root key
    print!("[Step 3] Deriving BIP32 root key ... ");
    io::stdout().flush().unwrap();
    let root = match Xpriv::new_master(Network::Regtest, &seed) {
        Ok(k) => k,
        Err(e) => fail(&e.to_string()),
    };
    let fingerprint = root.fingerprint(&secp).to_string();
    if fingerprint != EXPECT_FINGERPRINT {
        fail(&format!("fingerprint {} != {}", fingerprint, EXPECT_FINGERPRINT));
    }
    println!("✓ (fingerprint: {})", fingerprint);

    // Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
    print!("[Step 4] Deriving BIP44 Legacy P2PKH address ... ");
    io::stdout().flush().unwrap();
    let path_p2pkh = DerivationPath::from_str("m/44'/1'/0'/0/0").unwrap();
    let child_p2pkh = root.derive_priv(&secp, &path_p2pkh).unwrap();
    let pk_p2pkh = PublicKey::new(child_p2pkh.private_key.public_key(&secp));
    let addr_p2pkh = Address::p2pkh(pk_p2pkh, Network::Regtest);
    if addr_p2pkh.to_string() != EXPECT_P2PKH {
        fail(&format!("P2PKH {} != {}", addr_p2pkh, EXPECT_P2PKH));
    }
    println!("✓ ({})", addr_p2pkh);

    // Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
    print!("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ");
    io::stdout().flush().unwrap();
    let path_p2wpkh = DerivationPath::from_str("m/84'/1'/0'/0/0").unwrap();
    let child_p2wpkh = root.derive_priv(&secp, &path_p2wpkh).unwrap();
    let cpk_p2wpkh = CompressedPublicKey(child_p2wpkh.private_key.public_key(&secp));
    let addr_p2wpkh = Address::p2wpkh(&cpk_p2wpkh, Network::Regtest);
    if addr_p2wpkh.to_string() != EXPECT_P2WPKH {
        fail(&format!("P2WPKH {} != {}", addr_p2wpkh, EXPECT_P2WPKH));
    }
    println!("✓ ({})", addr_p2wpkh);

    // Step 6: Taproot P2TR (m/86'/1'/0'/0/0, BIP86 key-path tweak applied by p2tr)
    print!("[Step 6] Deriving BIP86 Taproot P2TR address ... ");
    io::stdout().flush().unwrap();
    let path_p2tr = DerivationPath::from_str("m/86'/1'/0'/0/0").unwrap();
    let child_p2tr = root.derive_priv(&secp, &path_p2tr).unwrap();
    let internal: UntweakedPublicKey = child_p2tr.private_key.public_key(&secp).x_only_public_key().0;
    let addr_p2tr = Address::p2tr(&secp, internal, None, Network::Regtest);
    if addr_p2tr.to_string() != EXPECT_P2TR {
        fail(&format!("P2TR {} != {}", addr_p2tr, EXPECT_P2TR));
    }
    println!("✓ ({})", addr_p2tr);

    // Step 7: Validate derived addresses with Bitcoin Core node
    print!("[Step 7] Validating addresses against Bitcoin Core node ... ");
    io::stdout().flush().unwrap();
    for addr in &[addr_p2pkh.to_string(), addr_p2wpkh.to_string(), addr_p2tr.to_string()] {
        let val = match node_rpc.call("validateaddress", vec![json!(addr)]) {
            Ok(v) => v,
            Err(e) => fail(&e),
        };
        let is_valid = val.get("isvalid").and_then(|v| v.as_bool()).unwrap_or(false);
        if !is_valid {
            fail(&format!("address {} reported invalid by node", addr));
        }
    }
    println!("✓ (all addresses valid)");

    println!("======================================================");
    println!("Result: PASS ✓");
    println!("======================================================");
    process::exit(0);
}

/// Minimal hex decoder — avoids pulling an extra crate just for the test-vector seed.
fn hex_decode(s: &str) -> Result<Vec<u8>, ()> {
    if s.len() % 2 != 0 {
        return Err(());
    }
    (0..s.len())
        .step_by(2)
        .map(|i| u8::from_str_radix(&s[i..i + 2], 16).map_err(|_| ()))
        .collect()
}
