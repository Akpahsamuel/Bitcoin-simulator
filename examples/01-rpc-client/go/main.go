package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Akpahsamuel/Bitcoin-simulator/examples/go/common"
)

func main() {
	fmt.Println("=== Lab 01: RPC Client (Go) ===")

	// Step 1: Bootstrap lab wallet and initial funds
	fmt.Print("[Step 1] Bootstrapping lab wallet and initial funds ... ")
	bootstrapRes, err := common.BootstrapLab(nil)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		fmt.Println("======================================================")
		fmt.Println("Result: FAIL ✗")
		fmt.Println("======================================================")
		os.Exit(1)
	}
	fmt.Println("✓")

	nodeRPC := bootstrapRes.RPC
	walletRPC := bootstrapRes.WalletRPC

	// Step 2: Query blockchain info via JSON-RPC
	fmt.Print("[Step 2] Querying getblockchaininfo via JSON-RPC ... ")
	chainInfoRaw, err := nodeRPC.Call("getblockchaininfo")
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var chainInfo struct {
		Chain  string `json:"chain"`
		Blocks int    `json:"blocks"`
	}
	if err := json.Unmarshal(chainInfoRaw, &chainInfo); err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	if chainInfo.Chain != "regtest" || chainInfo.Blocks < 101 {
		fmt.Printf("\n✗ FAILURE: unexpected chain %s or blocks %d\n", chainInfo.Chain, chainInfo.Blocks)
		os.Exit(1)
	}
	fmt.Printf("✓ (chain: %s, blocks: %d)\n", chainInfo.Chain, chainInfo.Blocks)

	// Step 3: Generate fresh address and mine 1 block
	fmt.Print("[Step 3] Generating fresh address and mining 1 block ... ")
	addrRaw, err := walletRPC.Call("getnewaddress", "lab01_test", "bech32m")
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var freshAddr string
	_ = json.Unmarshal(addrRaw, &freshAddr)

	blockHashesRaw, err := walletRPC.Call("generatetoaddress", 1, freshAddr)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var blockHashes []string
	_ = json.Unmarshal(blockHashesRaw, &blockHashes)
	if len(blockHashes) != 1 {
		fmt.Println("\n✗ FAILURE: expected 1 block hash")
		os.Exit(1)
	}

	blockCountRaw, _ := nodeRPC.Call("getblockcount")
	var newBlocks int
	_ = json.Unmarshal(blockCountRaw, &newBlocks)
	fmt.Printf("✓ (new height: %d, mined block: %s...)\n", newBlocks, blockHashes[0][:16])

	// Step 4: Query wallet balance
	fmt.Print("[Step 4] Querying wallet balance via getbalance ... ")
	balanceRaw, err := walletRPC.Call("getbalance")
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var balance float64
	_ = json.Unmarshal(balanceRaw, &balance)
	if balance <= 0 {
		fmt.Printf("\n✗ FAILURE: expected balance > 0, got %f\n", balance)
		os.Exit(1)
	}
	fmt.Printf("✓ (balance: %f BTC)\n", balance)

	// Step 5: Query unauthenticated REST API
	fmt.Print("[Step 5] Querying unauthenticated REST API (/rest/chaininfo.json) ... ")
	restRaw, err := nodeRPC.GetREST("chaininfo.json")
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var restInfo struct {
		Chain  string `json:"chain"`
		Blocks int    `json:"blocks"`
	}
	if err := json.Unmarshal(restRaw, &restInfo); err != nil || restInfo.Chain != "regtest" || restInfo.Blocks != newBlocks {
		fmt.Println("\n✗ FAILURE: REST info verification failed")
		os.Exit(1)
	}
	fmt.Println("✓")

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
