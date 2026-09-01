use reqwest::blocking::Client;
use serde_json::{json, Value};
use std::env;
use std::path::{Path, PathBuf};

pub struct SandboxConfig {
    pub rpc_url: String,
    pub rpc_user: String,
    pub rpc_password: String,
    pub rpc_wallet: String,
    pub zmq_rawblock: String,
    pub zmq_rawtx: String,
}

fn find_repo_root() -> PathBuf {
    let mut current = env::current_dir().unwrap_or_else(|_| PathBuf::from("."));
    loop {
        if current.join(".git").exists() || current.join(".env.example").exists() {
            return current;
        }
        if !current.pop() {
            break;
        }
    }
    PathBuf::from(".")
}

pub fn get_config() -> SandboxConfig {
    let repo_root = find_repo_root();
    let env_file = repo_root.join(".env");
    let env_example = repo_root.join(".env.example");

    if env_file.exists() {
        let _ = dotenvy::from_path(&env_file);
    } else if env_example.exists() {
        let _ = dotenvy::from_path(&env_example);
    }

    SandboxConfig {
        rpc_url: env::var("BITCOIN_RPC_URL").unwrap_or_else(|_| "http://127.0.0.1:18443".to_string()),
        rpc_user: env::var("BITCOIN_RPC_USER").unwrap_or_else(|_| "bitcoinrpc".to_string()),
        rpc_password: env::var("BITCOIN_RPC_PASSWORD").unwrap_or_else(|_| "bitcoinrpcpassword".to_string()),
        rpc_wallet: env::var("BITCOIN_RPC_WALLET").unwrap_or_else(|_| "lab".to_string()),
        zmq_rawblock: env::var("BITCOIN_ZMQ_RAWBLOCK").unwrap_or_else(|_| "tcp://127.0.0.1:28332".to_string()),
        zmq_rawtx: env::var("BITCOIN_ZMQ_RAWTX").unwrap_or_else(|_| "tcp://127.0.0.1:28333".to_string()),
    }
}

#[derive(Clone)]
pub struct BitcoinRPC {
    pub base_url: String,
    pub user: String,
    pub password: String,
    pub wallet: Option<String>,
    client: Client,
}

impl BitcoinRPC {
    pub fn new() -> Self {
        let config = get_config();
        Self {
            base_url: config.rpc_url.trim_end_matches('/').to_string(),
            user: config.rpc_user,
            password: config.rpc_password,
            wallet: None,
            client: Client::new(),
        }
    }

    pub fn with_wallet(wallet: &str) -> Self {
        let mut rpc = Self::new();
        rpc.wallet = Some(wallet.to_string());
        rpc
    }

    pub fn url(&self) -> String {
        if let Some(ref w) = self.wallet {
            format!("{}/wallet/{}", self.base_url, w)
        } else {
            self.base_url.clone()
        }
    }

    pub fn for_wallet(&self, wallet: &str) -> Self {
        let mut cloned = self.clone();
        cloned.wallet = Some(wallet.to_string());
        cloned
    }

    pub fn call(&self, method: &str, params: Vec<Value>) -> Result<Value, String> {
        let payload = json!({
            "jsonrpc": "1.0",
            "id": "bitcoin-sandbox-rs",
            "method": method,
            "params": params
        });

        let resp = self
            .client
            .post(self.url())
            .basic_auth(&self.user, Some(&self.password))
            .json(&payload)
            .send()
            .map_err(|e| format!("Could not connect to Bitcoin node at {}: {}", self.url(), e))?;

        let body: Value = resp
            .json()
            .map_err(|e| format!("Failed to parse response JSON: {}", e))?;

        if let Some(err) = body.get("error").filter(|e| !e.is_null()) {
            return Err(format!("Bitcoin RPC Error: {}", err));
        }

        Ok(body.get("result").cloned().unwrap_or(Value::Null))
    }

    pub fn get_rest(&self, endpoint: &str) -> Result<Value, String> {
        let clean = endpoint.trim_start_matches('/');
        let url = format!("{}/rest/{}", self.base_url, clean);

        let resp = self
            .client
            .get(&url)
            .send()
            .map_err(|e| format!("Could not connect to REST endpoint at {}: {}", url, e))?;

        if !resp.status().is_success() {
            return Err(format!("REST endpoint returned status: {}", resp.status()));
        }

        resp.json().map_err(|e| format!("Failed to parse REST JSON: {}", e))
    }
}
