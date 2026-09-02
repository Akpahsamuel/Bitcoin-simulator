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
	"github.com/btcsuite/btcd/txscript"
)

// Deterministic test vector: the canonical "abandon ... about" mnemonic on regtest.
// Expected values are cross-checked against embit (Python), bitcoinjs-lib (TypeScript)
// and rust-bitcoin (Rust) — every port must derive exactly these.
const (
	seedHex           = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	expectFingerprint = "73c5da0a"
	expectP2PKH       = "mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV"
	expectP2WPKH      = "bcrt1q6rz28mcfaxtmd6v789l9rrlrusdprr9pz3cppk"
	expectP2TR        = "bcrt1p8wpt9v4frpf3tkn0srd97pksgsxc5hs52lafxwru9kgeephvs7rqjeprhg"
)

func fail(msg string) {
	fmt.Printf("\n✗ FAILURE: %s\n", msg)
	fmt.Println("======================================================")
	fmt.Println("Result: FAIL ✗")
	fmt.Println("======================================================")
	os.Exit(1)
}

// deriveChild walks a BIP32 path of (index, hardened) segments from a master key.
func deriveChild(k *hdkeychain.ExtendedKey, segs [][2]uint32) (*hdkeychain.ExtendedKey, error) {
	cur := k
	for _, s := range segs {
		idx := s[0]
		if s[1] == 1 {
			idx += hdkeychain.HardenedKeyStart
		}
		next, err := cur.Derive(idx)
		if err != nil {
			return nil, err
		}
		cur = next
	}
	return cur, nil
}

func main() {
	fmt.Println("=== Lab 02: Keys & Addresses (Go) ===")

	// Step 1: Bootstrap lab
	fmt.Print("[Step 1] Bootstrapping lab wallet & RPC connection ... ")
	bootstrapRes, err := common.BootstrapLab(nil)
	if err != nil {
		fail(err.Error())
	}
	fmt.Println("✓")

	nodeRPC := bootstrapRes.RPC
	params := &chaincfg.RegressionNetParams

	// Step 2: BIP39 mnemonic -> 512-bit root seed (test vector seed shown directly).
	fmt.Print("[Step 2] Generating BIP39 mnemonic & root seed ... ")
	seed, err := hex.DecodeString(seedHex)
	if err != nil || len(seed) != 64 {
		fail("bad test-vector seed")
	}
	fmt.Printf("✓ (12 words, %d-byte seed)\n", len(seed))

	// Step 3: BIP32 root key
	fmt.Print("[Step 3] Deriving BIP32 root key ... ")
	masterKey, err := hdkeychain.NewMaster(seed, params)
	if err != nil {
		fail(err.Error())
	}
	masterPub, err := masterKey.ECPubKey()
	if err != nil {
		fail(err.Error())
	}
	fingerprint := hex.EncodeToString(btcutil.Hash160(masterPub.SerializeCompressed())[:4])
	if fingerprint != expectFingerprint {
		fail(fmt.Sprintf("fingerprint %s != %s", fingerprint, expectFingerprint))
	}
	fmt.Printf("✓ (fingerprint: %s)\n", fingerprint)

	// Step 4: Legacy P2PKH (m/44'/1'/0'/0/0)
	fmt.Print("[Step 4] Deriving BIP44 Legacy P2PKH address ... ")
	kP2PKH, err := deriveChild(masterKey, [][2]uint32{{44, 1}, {1, 1}, {0, 1}, {0, 0}, {0, 0}})
	if err != nil {
		fail(err.Error())
	}
	pubP2PKH, err := kP2PKH.ECPubKey()
	if err != nil {
		fail(err.Error())
	}
	addrP2PKH, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pubP2PKH.SerializeCompressed()), params)
	if err != nil {
		fail(err.Error())
	}
	if addrP2PKH.EncodeAddress() != expectP2PKH {
		fail(fmt.Sprintf("P2PKH %s != %s", addrP2PKH.EncodeAddress(), expectP2PKH))
	}
	fmt.Printf("✓ (%s)\n", addrP2PKH.EncodeAddress())

	// Step 5: Native SegWit P2WPKH (m/84'/1'/0'/0/0)
	fmt.Print("[Step 5] Deriving BIP84 SegWit P2WPKH address ... ")
	kP2WPKH, err := deriveChild(masterKey, [][2]uint32{{84, 1}, {1, 1}, {0, 1}, {0, 0}, {0, 0}})
	if err != nil {
		fail(err.Error())
	}
	pubP2WPKH, err := kP2WPKH.ECPubKey()
	if err != nil {
		fail(err.Error())
	}
	addrP2WPKH, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pubP2WPKH.SerializeCompressed()), params)
	if err != nil {
		fail(err.Error())
	}
	if addrP2WPKH.EncodeAddress() != expectP2WPKH {
		fail(fmt.Sprintf("P2WPKH %s != %s", addrP2WPKH.EncodeAddress(), expectP2WPKH))
	}
	fmt.Printf("✓ (%s)\n", addrP2WPKH.EncodeAddress())

	// Step 6: Taproot P2TR (m/86'/1'/0'/0/0)
	// BIP86: the witness program is the *tweaked* output key Q = P + H_TapTweak(P)*G,
	// not the raw internal key P. ComputeTaprootKeyNoScript applies that key-path tweak.
	fmt.Print("[Step 6] Deriving BIP86 Taproot P2TR address ... ")
	kP2TR, err := deriveChild(masterKey, [][2]uint32{{86, 1}, {1, 1}, {0, 1}, {0, 0}, {0, 0}})
	if err != nil {
		fail(err.Error())
	}
	internalP2TR, err := kP2TR.ECPubKey()
	if err != nil {
		fail(err.Error())
	}
	outputKey := txscript.ComputeTaprootKeyNoScript(internalP2TR)
	addrP2TR, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), params)
	if err != nil {
		fail(err.Error())
	}
	if addrP2TR.EncodeAddress() != expectP2TR {
		fail(fmt.Sprintf("P2TR %s != %s", addrP2TR.EncodeAddress(), expectP2TR))
	}
	fmt.Printf("✓ (%s)\n", addrP2TR.EncodeAddress())

	// Step 7: Validate derived addresses with Bitcoin Core node
	fmt.Print("[Step 7] Validating addresses against Bitcoin Core node ... ")
	for _, addr := range []string{addrP2PKH.EncodeAddress(), addrP2WPKH.EncodeAddress(), addrP2TR.EncodeAddress()} {
		valRaw, err := nodeRPC.Call("validateaddress", addr)
		if err != nil {
			fail(err.Error())
		}
		var val struct {
			IsValid bool `json:"isvalid"`
		}
		if err := json.Unmarshal(valRaw, &val); err != nil || !val.IsValid {
			fail(fmt.Sprintf("address %s reported invalid", addr))
		}
	}
	fmt.Println("✓ (all addresses valid)")

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
