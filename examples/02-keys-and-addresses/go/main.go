package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Akpahsamuel/Bitcoin-simulator/examples/go/common"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

func main() {
	fmt.Println("=== Lab 02: Keys & Addresses (Go) ===")

	// Step 1: Bootstrap lab
	fmt.Print("[Step 1] Bootstrapping lab wallet & RPC connection ... ")
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
	params := &chaincfg.RegressionNetParams

	// Step 2: Binary seed for 'abandon abandon ... about'
	fmt.Print("[Step 2] Generating BIP39 binary seed ... ")
	seedHex := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	seed, _ := hex.DecodeString(seedHex)
	fmt.Println("✓ (64 bytes)")

	// Step 3: BIP32 Root Key
	fmt.Print("[Step 3] Deriving BIP32 root key ... ")
	masterKey, err := hdkeychain.NewMaster(seed, params)
	if err != nil {
		fmt.Printf("\n✗ FAILURE: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ (fingerprint: %x)\n", masterKey.ParentFingerprint())

	// Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
	fmt.Print("[Step 4] Deriving BIP44 Legacy P2PKH address ... ")
	// Derive m/44'/1'/0'/0/0
	k44, _ := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	k44_1, _ := k44.Derive(hdkeychain.HardenedKeyStart + 1)
	k44_1_0, _ := k44_1.Derive(hdkeychain.HardenedKeyStart + 0)
	k44_1_0_0, _ := k44_1_0.Derive(0)
	kP2PKH, _ := k44_1_0_0.Derive(0)
	pubP2PKH, _ := kP2PKH.ECPubKey()
	addrP2PKH, _ := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubP2PKH.SerializeCompressed()), params)
	fmt.Printf("✓ (%s)\n", addrP2PKH.EncodeAddress())

	// Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
	fmt.Print("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ")
	k84, _ := masterKey.Derive(hdkeychain.HardenedKeyStart + 84)
	k84_1, _ := k84.Derive(hdkeychain.HardenedKeyStart + 1)
	k84_1_0, _ := k84_1.Derive(hdkeychain.HardenedKeyStart + 0)
	k84_1_0_0, _ := k84_1_0.Derive(0)
	kP2WPKH, _ := k84_1_0_0.Derive(0)
	pubP2WPKH, _ := kP2WPKH.ECPubKey()
	addrP2WPKH, _ := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubP2WPKH.SerializeCompressed()), params)
	fmt.Printf("✓ (%s)\n", addrP2WPKH.EncodeAddress())

	// Step 6: Taproot P2TR (m/86'/1'/0'/0/0)
	fmt.Print("[Step 6] Deriving BIP86 Taproot P2TR address ... ")
	k86, _ := masterKey.Derive(hdkeychain.HardenedKeyStart + 86)
	k86_1, _ := k86.Derive(hdkeychain.HardenedKeyStart + 1)
	k86_1_0, _ := k86_1.Derive(hdkeychain.HardenedKeyStart + 0)
	k86_1_0_0, _ := k86_1_0.Derive(0)
	kP2TR, _ := k86_1_0_0.Derive(0)
	pubP2TR, _ := kP2TR.ECPubKey()
	schnorrPubKey := schnorr.SerializePubKey(pubP2TR)
	addrP2TR, _ := btcutil.NewAddressTaproot(schnorrPubKey, params)
	fmt.Printf("✓ (%s)\n", addrP2TR.EncodeAddress())

	// Step 7: Validate derived addresses with Bitcoin Core node
	fmt.Print("[Step 7] Validating addresses against Bitcoin Core node ... ")
	for _, addr := range []string{addrP2PKH.EncodeAddress(), addrP2WPKH.EncodeAddress(), addrP2TR.EncodeAddress()} {
		valRaw, err := nodeRPC.Call("validateaddress", addr)
		if err != nil {
			fmt.Printf("\n✗ FAILURE: %v\n", err)
			os.Exit(1)
		}
		var val struct {
			IsValid bool `json:"isvalid"`
		}
		if err := json.Unmarshal(valRaw, &val); err != nil || !val.IsValid {
			fmt.Printf("\n✗ FAILURE: address %s reported invalid\n", addr)
			os.Exit(1)
		}
	}
	fmt.Println("✓ (all addresses valid)")

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
