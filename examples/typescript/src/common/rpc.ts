import * as fs from "node:fs";
import * as path from "node:path";
import { fileURLToPath } from "node:url";
import * as dotenv from "dotenv";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export function findRepoRoot(): string {
  let current = __dirname;
  while (current !== path.dirname(current)) {
    if (
      fs.existsSync(path.join(current, ".git")) ||
      fs.existsSync(path.join(current, ".env.example"))
    ) {
      return current;
    }
    current = path.dirname(current);
  }
  return process.cwd();
}

export interface SandboxConfig {
  rpcUrl: string;
  rpcUser: string;
  rpcPassword: string;
  rpcWallet: string;
  zmqRawBlock: string;
  zmqRawTx: string;
}

export function getConfig(): SandboxConfig {
  const repoRoot = findRepoRoot();
  const envPath = path.join(repoRoot, ".env");
  const envExamplePath = path.join(repoRoot, ".env.example");

  if (fs.existsSync(envPath)) {
    dotenv.config({ path: envPath });
  } else if (fs.existsSync(envExamplePath)) {
    dotenv.config({ path: envExamplePath });
  }

  return {
    rpcUrl: process.env.BITCOIN_RPC_URL || "http://127.0.0.1:18443",
    rpcUser: process.env.BITCOIN_RPC_USER || "bitcoinrpc",
    rpcPassword: process.env.BITCOIN_RPC_PASSWORD || "bitcoinrpcpassword",
    rpcWallet: process.env.BITCOIN_RPC_WALLET || "lab",
    zmqRawBlock: process.env.BITCOIN_ZMQ_RAWBLOCK || "tcp://127.0.0.1:28332",
    zmqRawTx: process.env.BITCOIN_ZMQ_RAWTX || "tcp://127.0.0.1:28333",
  };
}

export class BitcoinRPC {
  public baseUrl: string;
  public user: string;
  public password: string;
  public wallet?: string;

  constructor(options?: {
    url?: string;
    user?: string;
    password?: string;
    wallet?: string;
  }) {
    const config = getConfig();
    this.baseUrl = (options?.url || config.rpcUrl).replace(/\/+$/, "");
    this.user = options?.user || config.rpcUser;
    this.password = options?.password || config.rpcPassword;
    this.wallet = options?.wallet;
  }

  get url(): string {
    if (this.wallet) {
      return `${this.baseUrl}/wallet/${this.wallet}`;
    }
    return this.baseUrl;
  }

  async call<T = any>(method: string, params: any[] = []): Promise<T> {
    const authHeader =
      "Basic " + Buffer.from(`${this.user}:${this.password}`).toString("base64");

    let response: Response;
    try {
      response = await fetch(this.url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: authHeader,
        },
        body: JSON.stringify({
          jsonrpc: "1.0",
          id: "bitcoin-sandbox-ts",
          method,
          params,
        }),
      });
    } catch (err: any) {
      throw new Error(
        `Could not connect to Bitcoin node at ${this.url}. Is bitcoind running? (Run: bash scripts/init-lab.sh)`
      );
    }

    const data = (await response.json()) as any;
    if (data.error) {
      throw new Error(
        `Bitcoin RPC Error (${data.error.code}): ${data.error.message}`
      );
    }

    return data.result as T;
  }

  async getRest<T = any>(endpoint: string): Promise<T> {
    const cleanEndpoint = endpoint.replace(/^\/+/, "");
    const url = `${this.baseUrl}/rest/${cleanEndpoint}`;

    let response: Response;
    try {
      response = await fetch(url);
    } catch (err: any) {
      throw new Error(`Could not connect to Bitcoin REST endpoint at ${url}`);
    }

    if (!response.ok) {
      throw new Error(`REST request failed with status: ${response.status}`);
    }

    return (await response.json()) as T;
  }

  forWallet(walletName: string): BitcoinRPC {
    return new BitcoinRPC({
      url: this.baseUrl,
      user: this.user,
      password: this.password,
      wallet: walletName,
    });
  }
}
