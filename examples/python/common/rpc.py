"""Bitcoin JSON-RPC and REST client for Python examples."""
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional
import requests
from dotenv import load_dotenv


def find_repo_root() -> Path:
    """Locate the repository root directory by looking for .git or .env.example."""
    current = Path(__file__).resolve().parent
    while current != current.parent:
        if (current / ".git").exists() or (current / ".env.example").exists():
            return current
        current = current.parent
    return Path.cwd()


def get_config() -> Dict[str, str]:
    """Load configuration from environment or .env file."""
    repo_root = find_repo_root()
    env_path = repo_root / ".env"
    if env_path.exists():
        load_dotenv(dotenv_path=env_path)
    else:
        env_example = repo_root / ".env.example"
        if env_example.exists():
            load_dotenv(dotenv_path=env_example)

    return {
        "rpc_url": os.getenv("BITCOIN_RPC_URL", "http://127.0.0.1:18443"),
        "rpc_user": os.getenv("BITCOIN_RPC_USER", "bitcoinrpc"),
        "rpc_password": os.getenv("BITCOIN_RPC_PASSWORD", "bitcoinrpcpassword"),
        "rpc_wallet": os.getenv("BITCOIN_RPC_WALLET", "lab"),
        "zmq_rawblock": os.getenv("BITCOIN_ZMQ_RAWBLOCK", "tcp://127.0.0.1:28332"),
        "zmq_rawtx": os.getenv("BITCOIN_ZMQ_RAWTX", "tcp://127.0.0.1:28333"),
    }


class BitcoinRPC:
    """Client for Bitcoin Core JSON-RPC and REST endpoints."""

    def __init__(
        self,
        url: Optional[str] = None,
        user: Optional[str] = None,
        password: Optional[str] = None,
        wallet: Optional[str] = None,
    ):
        config = get_config()
        self.base_url = (url or config["rpc_url"]).rstrip("/")
        self.user = user or config["rpc_user"]
        self.password = password or config["rpc_password"]
        self.wallet = wallet
        self.session = requests.Session()
        self.session.auth = (self.user, self.password)

    @property
    def url(self) -> str:
        """Returns the full RPC endpoint URL including wallet context if set."""
        if self.wallet:
            return f"{self.base_url}/wallet/{self.wallet}"
        return self.base_url

    def call(self, method: str, params: Optional[List[Any]] = None) -> Any:
        """Execute a JSON-RPC method call."""
        if params is None:
            params = []
        payload = {
            "jsonrpc": "1.0",
            "id": "bitcoin-sandbox",
            "method": method,
            "params": params,
        }
        headers = {"content-type": "application/json"}
        try:
            response = self.session.post(self.url, json=payload, headers=headers, timeout=30)
        except requests.exceptions.ConnectionError as exc:
            raise ConnectionError(
                f"Could not connect to Bitcoin node at {self.url}. Is bitcoind running? (Run: bash scripts/init-lab.sh)"
            ) from exc

        if response.status_code not in (200, 500):
            response.raise_for_status()

        data = response.json()
        if data.get("error") is not None:
            err = data["error"]
            raise RuntimeError(f"Bitcoin RPC Error ({err.get('code')}): {err.get('message')}")

        return data.get("result")

    def get_rest(self, endpoint: str) -> Any:
        """Query an unauthenticated Bitcoin REST endpoint (e.g. 'chaininfo.json')."""
        clean_endpoint = endpoint.lstrip("/")
        url = f"{self.base_url}/rest/{clean_endpoint}"
        try:
            response = requests.get(url, timeout=10)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.ConnectionError as exc:
            raise ConnectionError(
                f"Could not connect to Bitcoin REST at {url}."
            ) from exc

    def for_wallet(self, wallet_name: str) -> "BitcoinRPC":
        """Return a new BitcoinRPC client scoped to a specific wallet."""
        return BitcoinRPC(
            url=self.base_url,
            user=self.user,
            password=self.password,
            wallet=wallet_name,
        )
