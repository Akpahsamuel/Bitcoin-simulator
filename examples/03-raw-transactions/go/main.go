package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/Akpahsamuel/Bitcoin-simulator/examples/go/common"
)

func main() {
	fmt.Println("=== Lab 03: Raw Transactions (Go) ===")

	// Step 1: Bootstrap lab
	fmt.Print("[Step 1] Bootstrapping lab wallet & ensure spendable UTXOs ... ")
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

	// Step 2: Select UTXO from listunspent
	fmt.Print("[Step 2] Selecting UTXO from listunspent ... ")
	unspentRaw, err := walletRPC.Call("listunspent", 1, 9999999)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var unspent []struct {
		TxID   string  `json:"txid"`
		Vout   uint32  `json:"vout"`
		Amount float64 `json:"amount"`
	}
	if err := json.Unmarshal(unspentRaw, &unspent); err != nil || len(unspent) == 0 {
		fmt.Println("\n✗ FAILURE: No spendable UTXOs found in wallet")
		os.Exit(1)
	}

	// Pick the largest UTXO — the shared `lab` wallet also holds small change
	// outputs from earlier lab runs, so the first entry is not reliably big enough.
	utxo := unspent[0]
	for _, u := range unspent[1:] {
		if u.Amount > utxo.Amount {
			utxo = u
		}
	}
	fmt.Printf("✓ (txid: %s..., vout: %d, amount: %f BTC)\n", utxo.TxID[:16], utxo.Vout, utxo.Amount)

	// Step 3: Construct raw transaction
	fmt.Print("[Step 3] Constructing raw transaction hex (createrawtransaction) ... ")
	recipientAddrRaw, _ := walletRPC.Call("getnewaddress", "recipient", "bech32m")
	var recipientAddr string
	_ = json.Unmarshal(recipientAddrRaw, &recipientAddr)

	changeAddrRaw, _ := walletRPC.Call("getnewaddress", "change", "bech32m")
	var changeAddr string
	_ = json.Unmarshal(changeAddrRaw, &changeAddr)

	sendAmount := 1.5
	fee := 0.0001
	changeAmount := math.Round((utxo.Amount-sendAmount-fee)*1e8) / 1e8
	if changeAmount <= 0 {
		fmt.Println("\n✗ FAILURE: UTXO amount insufficient for send + fee")
		os.Exit(1)
	}

	inputs := []map[string]interface{}{
		{"txid": utxo.TxID, "vout": utxo.Vout},
	}
	outputs := []map[string]interface{}{
		{recipientAddr: sendAmount},
		{changeAddr: changeAmount},
	}

	rawTxHexRaw, err := walletRPC.Call("createrawtransaction", inputs, outputs)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var rawTxHex string
	_ = json.Unmarshal(rawTxHexRaw, &rawTxHex)
	fmt.Printf("✓ (hex length: %d)\n", len(rawTxHex))

	// Step 4: Sign transaction inputs
	fmt.Print("[Step 4] Signing transaction inputs (signrawtransactionwithwallet) ... ")
	signRaw, err := walletRPC.Call("signrawtransactionwithwallet", rawTxHex)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var signResult struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	if err := json.Unmarshal(signRaw, &signResult); err != nil || !signResult.Complete {
		fmt.Println("\n✗ FAILURE: signing was incomplete")
		os.Exit(1)
	}
	fmt.Println("✓ (complete: true)")

	// Step 5: Verify transaction with testmempoolaccept
	fmt.Print("[Step 5] Verifying transaction with testmempoolaccept ... ")
	mempoolRaw, err := nodeRPC.Call("testmempoolaccept", []string{signResult.Hex})
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var mempoolRes []struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.Unmarshal(mempoolRaw, &mempoolRes); err != nil || len(mempoolRes) == 0 || !mempoolRes[0].Allowed {
		fmt.Println("\n✗ FAILURE: mempool rejected transaction")
		os.Exit(1)
	}
	fmt.Println("✓ (allowed: true)")

	// Step 6: Broadcast transaction via sendrawtransaction
	fmt.Print("[Step 6] Broadcasting transaction via sendrawtransaction ... ")
	broadcastTxidRaw, err := nodeRPC.Call("sendrawtransaction", signResult.Hex)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var broadcastTxid string
	_ = json.Unmarshal(broadcastTxidRaw, &broadcastTxid)
	fmt.Printf("✓ (txid: %s)\n", broadcastTxid)

	// Step 7: Mine 1 block & verify confirmation
	fmt.Print("[Step 7] Mining 1 block & verifying confirmation ... ")
	minerAddrRaw, _ := walletRPC.Call("getnewaddress", "miner", "bech32m")
	var minerAddr string
	_ = json.Unmarshal(minerAddrRaw, &minerAddr)
	_, _ = walletRPC.Call("generatetoaddress", 1, minerAddr)

	txDetailsRaw, err := walletRPC.Call("getrawtransaction", broadcastTxid, true)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	var txDetails struct {
		Confirmations int `json:"confirmations"`
	}
	if err := json.Unmarshal(txDetailsRaw, &txDetails); err != nil || txDetails.Confirmations < 1 {
		fmt.Println("\n✗ FAILURE: confirmations less than 1")
		os.Exit(1)
	}
	fmt.Printf("✓ (confirmations: %d)\n", txDetails.Confirmations)

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
