use crate::rpc::{get_config, BitcoinRPC};
use serde_json::{json, Value};

pub struct BootstrapResult {
    pub rpc: BitcoinRPC,
    pub wallet_rpc: BitcoinRPC,
}

pub fn bootstrap_lab(rpc_opt: Option<BitcoinRPC>) -> Result<BootstrapResult, String> {
    let config = get_config();
    let wallet_name = config.rpc_wallet;
    let node_rpc = rpc_opt.unwrap_or_else(BitcoinRPC::new);

    // 1. Check loaded wallets
    let loaded = node_rpc.call("listwallets", vec![])?;
    let is_loaded = loaded
        .as_array()
        .map(|arr| arr.iter().any(|v| v.as_str() == Some(&wallet_name)))
        .unwrap_or(false);

    if !is_loaded {
        let dir_info = node_rpc.call("listwalletdir", vec![])?;
        let exists_on_disk = dir_info
            .get("wallets")
            .and_then(|w| w.as_array())
            .map(|arr| {
                arr.iter().any(|w| {
                    w.get("name")
                        .and_then(|n| n.as_str())
                        == Some(&wallet_name)
                })
            })
            .unwrap_or(false);

        if exists_on_disk {
            node_rpc.call("loadwallet", vec![json!(wallet_name)])?;
        } else {
            node_rpc.call("createwallet", vec![json!(wallet_name)])?;
        }
    }

    let wallet_rpc = node_rpc.for_wallet(&wallet_name);

    // 2. Check balance and mine 101 blocks if 0
    let balance_val = wallet_rpc.call("getbalance", vec![])?;
    let balance = balance_val.as_f64().unwrap_or(0.0);

    if balance == 0.0 {
        let addr_val = wallet_rpc.call(
            "getnewaddress",
            vec![json!("bootstrap_mining"), json!("bech32m")],
        )?;
        let addr = addr_val
            .as_str()
            .ok_or_else(|| "Failed to get mining address".to_string())?;

        wallet_rpc.call("generatetoaddress", vec![json!(101), json!(addr)])?;
    }

    Ok(BootstrapResult {
        rpc: node_rpc,
        wallet_rpc,
    })
}
