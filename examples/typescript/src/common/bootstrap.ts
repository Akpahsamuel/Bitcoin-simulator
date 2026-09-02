import { BitcoinRPC, getConfig } from "./rpc.js";

export interface BootstrapResult {
  rpc: BitcoinRPC;
  walletRpc: BitcoinRPC;
}

export async function bootstrapLab(rpc?: BitcoinRPC): Promise<BootstrapResult> {
  const config = getConfig();
  const walletName = config.rpcWallet;
  const nodeRpc = rpc || new BitcoinRPC();

  // 1. Verify/Load or Create the lab wallet
  const loadedWallets = await nodeRpc.call<string[]>("listwallets");
  if (!loadedWallets.includes(walletName)) {
    const walletDirInfo = await nodeRpc.call<{ wallets: { name: string }[] }>("listwalletdir");
    const existingWallets = (walletDirInfo.wallets || []).map((w) => w.name);
    if (existingWallets.includes(walletName)) {
      await nodeRpc.call("loadwallet", [walletName]);
    } else {
      await nodeRpc.call("createwallet", [walletName]);
    }
  }

  const walletRpc = nodeRpc.forWallet(walletName);

  // 2. Check spendable balance; mine 101 blocks if 0 (coinbase maturity)
  const balance = await walletRpc.call<number>("getbalance");
  if (balance === 0) {
    const miningAddr = await walletRpc.call<string>("getnewaddress", ["bootstrap_mining", "bech32m"]);
    await walletRpc.call("generatetoaddress", [101, miningAddr]);
  }

  return { rpc: nodeRpc, walletRpc };
}
