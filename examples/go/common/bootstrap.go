package common

import (
	"encoding/json"
	"fmt"
)

type BootstrapResult struct {
	RPC       *BitcoinRPC
	WalletRPC *BitcoinRPC
}

func BootstrapLab(nodeRPC *BitcoinRPC) (*BootstrapResult, error) {
	if nodeRPC == nil {
		nodeRPC = NewBitcoinRPC()
	}
	cfg := GetConfig()
	walletName := cfg.RPCWallet

	// 1. Check loaded wallets
	loadedRaw, err := nodeRPC.Call("listwallets")
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}
	var loaded []string
	if err := json.Unmarshal(loadedRaw, &loaded); err != nil {
		return nil, fmt.Errorf("failed to parse listwallets response: %w", err)
	}

	isLoaded := false
	for _, w := range loaded {
		if w == walletName {
			isLoaded = true
			break
		}
	}

	if !isLoaded {
		dirRaw, err := nodeRPC.Call("listwalletdir")
		if err != nil {
			return nil, fmt.Errorf("failed to list wallet dir: %w", err)
		}
		var dirInfo struct {
			Wallets []struct {
				Name string `json:"name"`
			} `json:"wallets"`
		}
		_ = json.Unmarshal(dirRaw, &dirInfo)

		existsOnDisk := false
		for _, w := range dirInfo.Wallets {
			if w.Name == walletName {
				existsOnDisk = true
				break
			}
		}

		if existsOnDisk {
			if _, err := nodeRPC.Call("loadwallet", walletName); err != nil {
				return nil, fmt.Errorf("failed to load wallet %s: %w", walletName, err)
			}
		} else {
			if _, err := nodeRPC.Call("createwallet", walletName); err != nil {
				return nil, fmt.Errorf("failed to create wallet %s: %w", walletName, err)
			}
		}
	}

	walletRPC := nodeRPC.ForWallet(walletName)

	// 2. Check balance and mine 101 blocks if 0
	balanceRaw, err := walletRPC.Call("getbalance")
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	var balance float64
	if err := json.Unmarshal(balanceRaw, &balance); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	if balance == 0 {
		addrRaw, err := walletRPC.Call("getnewaddress", "bootstrap_mining", "bech32m")
		if err != nil {
			return nil, fmt.Errorf("failed to get mining address: %w", err)
		}
		var addr string
		if err := json.Unmarshal(addrRaw, &addr); err != nil {
			return nil, fmt.Errorf("failed to parse mining address: %w", err)
		}

		if _, err := walletRPC.Call("generatetoaddress", 101, addr); err != nil {
			return nil, fmt.Errorf("failed to mine 101 blocks: %w", err)
		}
	}

	return &BootstrapResult{
		RPC:       nodeRPC,
		WalletRPC: walletRPC,
	}, nil
}
