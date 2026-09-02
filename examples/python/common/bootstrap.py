"""Bootstrap helper for Python examples in the Bitcoin Sandbox."""
from typing import Tuple
from .rpc import BitcoinRPC, get_config


def bootstrap_lab(rpc: BitcoinRPC = None) -> Tuple[BitcoinRPC, BitcoinRPC]:
    """
    Ensures the 'lab' wallet exists, is loaded, and has spendable funds
    (by mining 101 blocks to clear coinbase maturity if necessary).
    
    Returns:
        (node_rpc, wallet_rpc): Pair of RPC clients for node-level and wallet-level calls.
    """
    config = get_config()
    wallet_name = config["rpc_wallet"]
    
    if rpc is None:
        rpc = BitcoinRPC()

    # 1. Verify/Load or Create the lab wallet
    loaded_wallets = rpc.call("listwallets")
    if wallet_name not in loaded_wallets:
        # Check if wallet directory exists on disk
        wallet_dir_info = rpc.call("listwalletdir")
        existing_wallets = [w["name"] for w in wallet_dir_info.get("wallets", [])]
        if wallet_name in existing_wallets:
            rpc.call("loadwallet", [wallet_name])
        else:
            # Create descriptor wallet
            rpc.call("createwallet", [wallet_name])

    wallet_rpc = rpc.for_wallet(wallet_name)

    # 2. Check spendable balance; mine 101 blocks if 0 (coinbase maturity)
    balance = wallet_rpc.call("getbalance")
    if balance == 0:
        mining_addr = wallet_rpc.call("getnewaddress", ["bootstrap_mining", "bech32m"])
        wallet_rpc.call("generatetoaddress", [101, mining_addr])

    return rpc, wallet_rpc
