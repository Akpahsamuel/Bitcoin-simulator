// Lab 04: Multisig Escrow (Go Reference Implementation)
// 2-of-3 P2WSH built from raw script, funded from the lab wallet, then spent by
// assembling the witness stack [<empty>, sigA, sigB, witnessScript] by hand.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/Akpahsamuel/Bitcoin-simulator/examples/go/common"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var skHex = [3]string{
	"1111111111111111111111111111111111111111111111111111111111111111",
	"2222222222222222222222222222222222222222222222222222222222222222",
	"3333333333333333333333333333333333333333333333333333333333333333",
}

const (
	expectWitnessScript = "5221034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aa2102466d7fcae563e5cb09a0d1870bb580344804617879a14949cf22285f1bae3f2721023c72addb4fdf09af94f0c94d7fe92a386a7e70cf8a1d85916386bb2535c7b1b153ae"
	expectP2WSH         = "bcrt1qpy8yjjs2l5neewx722mxve9w6m77zqsu7rldukggseflhwralerqh6ma0d"
	fundBTC             = 0.5
	feeSat              = 20_000
)

func fail(msg string) {
	fmt.Printf("\n✗ FAILURE: %s\n", msg)
	fmt.Println("======================================================")
	fmt.Println("Result: FAIL ✗")
	fmt.Println("======================================================")
	os.Exit(1)
}

func main() {
	fmt.Println("=== Lab 04: Multisig Escrow (Go) ===")
	params := &chaincfg.RegressionNetParams

	// Step 1
	fmt.Print("[Step 1] Bootstrapping lab wallet & funds ... ")
	res, err := common.BootstrapLab(nil)
	if err != nil {
		fail(err.Error())
	}
	nodeRPC, walletRPC := res.RPC, res.WalletRPC
	fmt.Println("✓")

	// Step 2
	fmt.Print("[Step 2] Deriving 3 escrow keypairs from test vectors ... ")
	var privs [3]*btcec.PrivateKey
	var pubs [3][]byte
	for i, h := range skHex {
		b, err := hex.DecodeString(h)
		if err != nil {
			fail("bad test secret")
		}
		privs[i], _ = btcec.PrivKeyFromBytes(b)
		pubs[i] = privs[i].PubKey().SerializeCompressed()
	}
	fmt.Printf("✓ (%s…, %s…, %s…)\n", hex.EncodeToString(pubs[0])[:10], hex.EncodeToString(pubs[1])[:10], hex.EncodeToString(pubs[2])[:10])

	// Step 3
	fmt.Print("[Step 3] Building 2-of-3 witness script ... ")
	witnessScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_2).
		AddData(pubs[0]).AddData(pubs[1]).AddData(pubs[2]).
		AddOp(txscript.OP_3).
		AddOp(txscript.OP_CHECKMULTISIG).
		Script()
	if err != nil {
		fail(err.Error())
	}
	if hex.EncodeToString(witnessScript) != expectWitnessScript {
		fail("witness script != canonical")
	}
	fmt.Println("✓ (matches canonical)")

	// Step 4
	fmt.Print("[Step 4] Deriving P2WSH address & validating with node ... ")
	wsHash := sha256.Sum256(witnessScript)
	p2wshAddr, err := btcutil.NewAddressWitnessScriptHash(wsHash[:], params)
	if err != nil {
		fail(err.Error())
	}
	if p2wshAddr.EncodeAddress() != expectP2WSH {
		fail(fmt.Sprintf("address %s != %s", p2wshAddr.EncodeAddress(), expectP2WSH))
	}
	valRaw, err := nodeRPC.Call("validateaddress", p2wshAddr.EncodeAddress())
	if err != nil {
		fail(err.Error())
	}
	var val struct {
		IsValid bool `json:"isvalid"`
	}
	if json.Unmarshal(valRaw, &val); !val.IsValid {
		fail("node rejected the P2WSH address")
	}
	fmt.Printf("✓ (%s…)\n", p2wshAddr.EncodeAddress()[:14])

	// Step 5
	fmt.Printf("[Step 5] Funding the multisig (%.1f BTC) & mining 1 block ... ", fundBTC)
	fundTxidRaw, err := walletRPC.Call("sendtoaddress", p2wshAddr.EncodeAddress(), fundBTC)
	if err != nil {
		fail(err.Error())
	}
	var fundTxid string
	_ = json.Unmarshal(fundTxidRaw, &fundTxid)
	minerRaw, _ := walletRPC.Call("getnewaddress", "lab04_miner", "bech32m")
	var miner string
	_ = json.Unmarshal(minerRaw, &miner)
	if _, err := walletRPC.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✓ (txid: %s…)\n", fundTxid[:16])

	// Step 6
	fmt.Print("[Step 6] Locating the funding UTXO by scriptPubKey ... ")
	p2wshPkScript, err := txscript.PayToAddrScript(p2wshAddr)
	if err != nil {
		fail(err.Error())
	}
	spkHex := hex.EncodeToString(p2wshPkScript)
	fundTxRaw, err := nodeRPC.Call("getrawtransaction", fundTxid, true)
	if err != nil {
		fail(err.Error())
	}
	var fundTx struct {
		Vout []struct {
			N            uint32  `json:"n"`
			Value        float64 `json:"value"`
			ScriptPubKey struct {
				Hex string `json:"hex"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
	}
	if err := json.Unmarshal(fundTxRaw, &fundTx); err != nil {
		fail(err.Error())
	}
	var fundVout uint32
	var fundSat int64
	found := false
	for _, o := range fundTx.Vout {
		if o.ScriptPubKey.Hex == spkHex {
			fundVout = o.N
			fundSat = int64(math.Round(o.Value * 1e8))
			found = true
			break
		}
	}
	if !found {
		fail("funding output not found")
	}
	fmt.Printf("✓ (vout: %d, %.8f BTC)\n", fundVout, float64(fundSat)/1e8)

	// Step 7
	fmt.Print("[Step 7] Building the spend transaction ... ")
	returnAddrRaw, _ := walletRPC.Call("getnewaddress", "escrow_return", "bech32m")
	var returnAddrStr string
	_ = json.Unmarshal(returnAddrRaw, &returnAddrStr)
	returnAddr, err := btcutil.DecodeAddress(returnAddrStr, params)
	if err != nil {
		fail(err.Error())
	}
	returnPkScript, err := txscript.PayToAddrScript(returnAddr)
	if err != nil {
		fail(err.Error())
	}
	outSat := fundSat - feeSat
	if outSat <= 0 {
		fail("funded amount too small for fee")
	}
	tx := wire.NewMsgTx(2)
	prevHash, err := chainhash.NewHashFromStr(fundTxid)
	if err != nil {
		fail(err.Error())
	}
	tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(prevHash, fundVout), nil, nil))
	tx.AddTxOut(wire.NewTxOut(outSat, returnPkScript))
	fmt.Println("✓")

	// Step 8
	fmt.Print("[Step 8] BIP143 sighash + signing with keys A and B ... ")
	fetcher := txscript.NewCannedPrevOutputFetcher(p2wshPkScript, fundSat)
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)
	sigHash, err := txscript.CalcWitnessSigHash(witnessScript, sigHashes, txscript.SigHashAll, tx, 0, fundSat)
	if err != nil {
		fail(err.Error())
	}
	sign := func(k *btcec.PrivateKey) []byte {
		return append(ecdsa.Sign(k, sigHash).Serialize(), byte(txscript.SigHashAll))
	}
	sigA := sign(privs[0])
	sigB := sign(privs[1])
	fmt.Printf("✓ (sigA %dB, sigB %dB)\n", len(sigA), len(sigB))

	// Step 9
	fmt.Print("[Step 9] Assembling witness [<empty>, sigA, sigB, script] ... ")
	tx.TxIn[0].Witness = wire.TxWitness{[]byte{}, sigA, sigB, witnessScript}
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		fail(err.Error())
	}
	rawHex := hex.EncodeToString(buf.Bytes())
	fmt.Printf("✓ (tx %d bytes)\n", len(buf.Bytes()))

	// Step 10
	fmt.Print("[Step 10] testmempoolaccept ... ")
	acceptRaw, err := nodeRPC.Call("testmempoolaccept", []string{rawHex})
	if err != nil {
		fail(err.Error())
	}
	var accept []struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.Unmarshal(acceptRaw, &accept); err != nil || len(accept) == 0 || !accept[0].Allowed {
		fail(fmt.Sprintf("mempool rejected: %s", string(acceptRaw)))
	}
	fmt.Println("✓ (allowed: true)")

	// Step 11
	fmt.Print("[Step 11] Broadcasting & mining 1 block ... ")
	spendTxidRaw, err := nodeRPC.Call("sendrawtransaction", rawHex)
	if err != nil {
		fail(err.Error())
	}
	var spendTxid string
	_ = json.Unmarshal(spendTxidRaw, &spendTxid)
	if _, err := walletRPC.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	detailsRaw, err := nodeRPC.Call("getrawtransaction", spendTxid, true)
	if err != nil {
		fail(err.Error())
	}
	var details struct {
		Confirmations int `json:"confirmations"`
	}
	if err := json.Unmarshal(detailsRaw, &details); err != nil || details.Confirmations < 1 {
		fail(fmt.Sprintf("expected confirmations >= 1, got %d", details.Confirmations))
	}
	fmt.Printf("✓ (confirmations: %d)\n", details.Confirmations)

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
