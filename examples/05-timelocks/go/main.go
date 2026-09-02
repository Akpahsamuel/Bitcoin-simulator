// Lab 05: Timelocks (Go Reference Implementation)
// CLTV (absolute, BIP65) and CSV (relative, BIP112) P2WSH outputs: build the
// script, fund it, watch testmempoolaccept reject an early spend, advance the
// chain, then spend successfully.
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

const (
	skAHex          = "1111111111111111111111111111111111111111111111111111111111111111"
	expectCSVScript = "53b27521034f355bdcb7cc0af728ef3cceb9615d90684bb5b2ca5f859ab0f0b704075871aaac"
	expectCSVAddr   = "bcrt1ql739fkda7sf20qkdwgku2j0ppeff4r7vsqasvvxestsqwtvuak3s9rktmg"
	csvDelay        = 3
	fundBTC         = 0.2
	feeSat          = 20_000
)

var params = &chaincfg.RegressionNetParams

func fail(msg string) {
	fmt.Printf("\n✗ FAILURE: %s\n", msg)
	fmt.Println("======================================================")
	fmt.Println("Result: FAIL ✗")
	fmt.Println("======================================================")
	os.Exit(1)
}

func str(raw json.RawMessage) string {
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

func i64(raw json.RawMessage) int64 {
	var n int64
	_ = json.Unmarshal(raw, &n)
	return n
}

func p2wshAddress(script []byte) *btcutil.AddressWitnessScriptHash {
	h := sha256.Sum256(script)
	a, err := btcutil.NewAddressWitnessScriptHash(h[:], params)
	if err != nil {
		fail(err.Error())
	}
	return a
}

func findVout(node *common.BitcoinRPC, txid, spkHex string) (uint32, int64) {
	raw, err := node.Call("getrawtransaction", txid, true)
	if err != nil {
		fail(err.Error())
	}
	var tx struct {
		Vout []struct {
			N            uint32  `json:"n"`
			Value        float64 `json:"value"`
			ScriptPubKey struct {
				Hex string `json:"hex"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
	}
	if err := json.Unmarshal(raw, &tx); err != nil {
		fail(err.Error())
	}
	for _, o := range tx.Vout {
		if o.ScriptPubKey.Hex == spkHex {
			return o.N, int64(math.Round(o.Value * 1e8))
		}
	}
	fail("funding output not found")
	return 0, 0
}

func buildSpend(node, wallet *common.BitcoinRPC, priv *btcec.PrivateKey, witnessScript []byte, txid string, vout uint32, valueSat int64, version int32, sequence uint32, lockTime uint32) string {
	destRaw, _ := wallet.Call("getnewaddress", "timelock_return", "bech32m")
	dest, err := btcutil.DecodeAddress(str(destRaw), params)
	if err != nil {
		fail(err.Error())
	}
	destPk, err := txscript.PayToAddrScript(dest)
	if err != nil {
		fail(err.Error())
	}
	outSat := valueSat - feeSat
	if outSat <= 0 {
		fail("funded amount too small for fee")
	}
	tx := wire.NewMsgTx(version)
	tx.LockTime = lockTime
	prevHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		fail(err.Error())
	}
	txin := wire.NewTxIn(wire.NewOutPoint(prevHash, vout), nil, nil)
	txin.Sequence = sequence
	tx.AddTxIn(txin)
	tx.AddTxOut(wire.NewTxOut(outSat, destPk))

	fetcher := txscript.NewCannedPrevOutputFetcher(mustP2WSHScript(witnessScript), valueSat)
	sigHashes := txscript.NewTxSigHashes(tx, fetcher)
	h, err := txscript.CalcWitnessSigHash(witnessScript, sigHashes, txscript.SigHashAll, tx, 0, valueSat)
	if err != nil {
		fail(err.Error())
	}
	sig := append(ecdsa.Sign(priv, h).Serialize(), byte(txscript.SigHashAll))
	tx.TxIn[0].Witness = wire.TxWitness{sig, witnessScript}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		fail(err.Error())
	}
	_ = node
	return hex.EncodeToString(buf.Bytes())
}

func mustP2WSHScript(witnessScript []byte) []byte {
	pk, err := txscript.PayToAddrScript(p2wshAddress(witnessScript))
	if err != nil {
		fail(err.Error())
	}
	return pk
}

func mempoolAllowed(node *common.BitcoinRPC, rawHex string) (bool, string) {
	raw, err := node.Call("testmempoolaccept", []string{rawHex})
	if err != nil {
		fail(err.Error())
	}
	var res []struct {
		Allowed      bool   `json:"allowed"`
		RejectReason string `json:"reject-reason"`
	}
	if err := json.Unmarshal(raw, &res); err != nil || len(res) == 0 {
		fail("bad testmempoolaccept response")
	}
	return res[0].Allowed, res[0].RejectReason
}

func main() {
	fmt.Println("=== Lab 05: Timelocks (Go) ===")

	fmt.Print("[Step 1] Bootstrapping lab wallet & funds ... ")
	res, err := common.BootstrapLab(nil)
	if err != nil {
		fail(err.Error())
	}
	node, wallet := res.RPC, res.WalletRPC
	skBytes, _ := hex.DecodeString(skAHex)
	priv, _ := btcec.PrivKeyFromBytes(skBytes)
	pubA := priv.PubKey().SerializeCompressed()
	minerRaw, _ := wallet.Call("getnewaddress", "lab05_miner", "bech32m")
	miner := str(minerRaw)
	fmt.Println("✓")

	// ---------- CLTV ----------
	fmt.Print("[Step 2] CLTV: building <height> OP_CLTV script ... ")
	heightRaw, _ := node.Call("getblockcount")
	lockHeight := i64(heightRaw) + 10
	cltvScript, err := txscript.NewScriptBuilder().
		AddInt64(lockHeight).
		AddOp(txscript.OP_CHECKLOCKTIMEVERIFY).AddOp(txscript.OP_DROP).
		AddData(pubA).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		fail(err.Error())
	}
	cltvAddr := p2wshAddress(cltvScript)
	fmt.Printf("✓ (lock at height %d)\n", lockHeight)

	fmt.Print("[Step 3] CLTV: deriving P2WSH address & validating with node ... ")
	valRaw, _ := node.Call("validateaddress", cltvAddr.EncodeAddress())
	var val struct {
		IsValid bool `json:"isvalid"`
	}
	if json.Unmarshal(valRaw, &val); !val.IsValid {
		fail("node rejected CLTV address")
	}
	fmt.Printf("✓ (%s…)\n", cltvAddr.EncodeAddress()[:14])

	fmt.Printf("[Step 4] CLTV: funding (%.1f BTC) & mining 1 block ... ", fundBTC)
	cltvTxidRaw, err := wallet.Call("sendtoaddress", cltvAddr.EncodeAddress(), fundBTC)
	if err != nil {
		fail(err.Error())
	}
	cltvTxid := str(cltvTxidRaw)
	if _, err := wallet.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	cltvSpk := hex.EncodeToString(mustP2WSHScript(cltvScript))
	cltvVout, cltvSat := findVout(node, cltvTxid, cltvSpk)
	fmt.Printf("✓ (vout %d)\n", cltvVout)

	fmt.Print("[Step 5] CLTV: early spend rejected by testmempoolaccept ... ")
	raw := buildSpend(node, wallet, priv, cltvScript, cltvTxid, cltvVout, cltvSat, 2, 0xFFFFFFFE, uint32(lockHeight))
	allowed, reason := mempoolAllowed(node, raw)
	if allowed {
		fail("expected the early CLTV spend to be rejected")
	}
	fmt.Printf("✓ (allowed: false — %s)\n", reason)

	fmt.Print("[Step 6] CLTV: mining to the lock height ... ")
	curRaw, _ := node.Call("getblockcount")
	if need := lockHeight + 1 - i64(curRaw); need > 0 {
		if _, err := wallet.Call("generatetoaddress", need, miner); err != nil {
			fail(err.Error())
		}
	}
	newHRaw, _ := node.Call("getblockcount")
	fmt.Printf("✓ (height %d)\n", i64(newHRaw))

	fmt.Print("[Step 7] CLTV: spend accepted, broadcast & confirmed ... ")
	raw = buildSpend(node, wallet, priv, cltvScript, cltvTxid, cltvVout, cltvSat, 2, 0xFFFFFFFE, uint32(lockHeight))
	if allowed, _ := mempoolAllowed(node, raw); !allowed {
		fail("mempool rejected the mature CLTV spend")
	}
	spRaw, err := node.Call("sendrawtransaction", raw)
	if err != nil {
		fail(err.Error())
	}
	if _, err := wallet.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	if confirmations(node, str(spRaw)) < 1 {
		fail("CLTV spend not confirmed")
	}
	fmt.Printf("✓ (confirmations: %d)\n", confirmations(node, str(spRaw)))

	// ---------- CSV ----------
	fmt.Print("[Step 8] CSV: building <3> OP_CSV script ... ")
	csvScript, err := txscript.NewScriptBuilder().
		AddInt64(csvDelay).
		AddOp(txscript.OP_CHECKSEQUENCEVERIFY).AddOp(txscript.OP_DROP).
		AddData(pubA).AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		fail(err.Error())
	}
	if hex.EncodeToString(csvScript) != expectCSVScript {
		fail("CSV script != canonical")
	}
	csvAddr := p2wshAddress(csvScript)
	fmt.Println("✓ (matches canonical)")

	fmt.Print("[Step 9] CSV: deriving P2WSH address & validating with node ... ")
	if csvAddr.EncodeAddress() != expectCSVAddr {
		fail(fmt.Sprintf("%s != %s", csvAddr.EncodeAddress(), expectCSVAddr))
	}
	valRaw, _ = node.Call("validateaddress", csvAddr.EncodeAddress())
	if json.Unmarshal(valRaw, &val); !val.IsValid {
		fail("node rejected CSV address")
	}
	fmt.Printf("✓ (%s…)\n", csvAddr.EncodeAddress()[:14])

	fmt.Printf("[Step 10] CSV: funding (%.1f BTC) & mining 1 block ... ", fundBTC)
	csvTxidRaw, err := wallet.Call("sendtoaddress", csvAddr.EncodeAddress(), fundBTC)
	if err != nil {
		fail(err.Error())
	}
	csvTxid := str(csvTxidRaw)
	if _, err := wallet.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	csvSpk := hex.EncodeToString(mustP2WSHScript(csvScript))
	csvVout, csvSat := findVout(node, csvTxid, csvSpk)
	fmt.Printf("✓ (vout %d, 1 confirmation)\n", csvVout)

	fmt.Print("[Step 11] CSV: early spend rejected by testmempoolaccept ... ")
	raw = buildSpend(node, wallet, priv, csvScript, csvTxid, csvVout, csvSat, 2, csvDelay, 0)
	if allowed, reason := mempoolAllowed(node, raw); allowed {
		fail("expected the early CSV spend to be rejected")
	} else {
		fmt.Printf("✓ (allowed: false — %s)\n", reason)
	}

	fmt.Print("[Step 12] CSV: mining 3 blocks to satisfy the relative delay ... ")
	if _, err := wallet.Call("generatetoaddress", csvDelay, miner); err != nil {
		fail(err.Error())
	}
	fmt.Println("✓")

	fmt.Print("[Step 13] CSV: spend accepted, broadcast & confirmed ... ")
	raw = buildSpend(node, wallet, priv, csvScript, csvTxid, csvVout, csvSat, 2, csvDelay, 0)
	if allowed, _ := mempoolAllowed(node, raw); !allowed {
		fail("mempool rejected the mature CSV spend")
	}
	spRaw, err = node.Call("sendrawtransaction", raw)
	if err != nil {
		fail(err.Error())
	}
	if _, err := wallet.Call("generatetoaddress", 1, miner); err != nil {
		fail(err.Error())
	}
	c := confirmations(node, str(spRaw))
	if c < 1 {
		fail("CSV spend not confirmed")
	}
	fmt.Printf("✓ (confirmations: %d)\n", c)

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}

func confirmations(node *common.BitcoinRPC, txid string) int {
	raw, err := node.Call("getrawtransaction", txid, true)
	if err != nil {
		return 0
	}
	var d struct {
		Confirmations int `json:"confirmations"`
	}
	_ = json.Unmarshal(raw, &d)
	return d.Confirmations
}
