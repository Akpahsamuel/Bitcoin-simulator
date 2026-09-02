package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	RPCURL      string
	RPCUser     string
	RPCPassword string
	RPCWallet   string
	ZMQRawBlock string
	ZMQRawTx    string
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".env.example")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func GetConfig() Config {
	repoRoot := findRepoRoot()
	envPath := filepath.Join(repoRoot, ".env")
	envExamplePath := filepath.Join(repoRoot, ".env.example")

	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load(envPath)
	} else if _, err := os.Stat(envExamplePath); err == nil {
		_ = godotenv.Load(envExamplePath)
	}

	getEnv := func(key, defaultVal string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return defaultVal
	}

	return Config{
		RPCURL:      getEnv("BITCOIN_RPC_URL", "http://127.0.0.1:18443"),
		RPCUser:     getEnv("BITCOIN_RPC_USER", "bitcoinrpc"),
		RPCPassword: getEnv("BITCOIN_RPC_PASSWORD", "bitcoinrpcpassword"),
		RPCWallet:   getEnv("BITCOIN_RPC_WALLET", "lab"),
		ZMQRawBlock: getEnv("BITCOIN_ZMQ_RAWBLOCK", "tcp://127.0.0.1:28332"),
		ZMQRawTx:    getEnv("BITCOIN_ZMQ_RAWTX", "tcp://127.0.0.1:28333"),
	}
}

type BitcoinRPC struct {
	BaseURL  string
	User     string
	Password string
	Wallet   string
	client   *http.Client
}

func NewBitcoinRPC() *BitcoinRPC {
	cfg := GetConfig()
	return &BitcoinRPC{
		BaseURL:  strings.TrimRight(cfg.RPCURL, "/"),
		User:     cfg.RPCUser,
		Password: cfg.RPCPassword,
		client:   &http.Client{},
	}
}

func (r *BitcoinRPC) ForWallet(walletName string) *BitcoinRPC {
	return &BitcoinRPC{
		BaseURL:  r.BaseURL,
		User:     r.User,
		Password: r.Password,
		Wallet:   walletName,
		client:   r.client,
	}
}

func (r *BitcoinRPC) URL() string {
	if r.Wallet != "" {
		return fmt.Sprintf("%s/wallet/%s", r.BaseURL, r.Wallet)
	}
	return r.BaseURL
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID string `json:"id"`
}

func (r *BitcoinRPC) Call(method string, params ...interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}
	reqBody, err := json.Marshal(rpcRequest{
		JSONRPC: "1.0",
		ID:      "bitcoin-sandbox-go",
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	req, err := http.NewRequest("POST", r.URL(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.User, r.Password)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not connect to Bitcoin node at %s: %w", r.URL(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("bitcoin RPC Error (%d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func (r *BitcoinRPC) GetREST(endpoint string) (json.RawMessage, error) {
	clean := strings.TrimLeft(endpoint, "/")
	url := fmt.Sprintf("%s/rest/%s", r.BaseURL, clean)

	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not connect to REST endpoint at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST request failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
