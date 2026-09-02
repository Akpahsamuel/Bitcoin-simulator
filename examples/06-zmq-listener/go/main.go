// Lab 06: ZMQ Listener (Go Reference Implementation)
// Subscribe to Bitcoin Core's rawtx / rawblock ZeroMQ feeds, trigger one
// transaction and one block, and verify the pushed bytes against the node.
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Akpahsamuel/Bitcoin-simulator/examples/go/common"
	zmq "github.com/pebbe/zmq4"
)

const recvTimeout = 15 * time.Second

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

func subscribe(endpoint, topic string) *zmq.Socket {
	sock, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		fail(err.Error())
	}
	if err := sock.SetSubscribe(topic); err != nil {
		fail(err.Error())
	}
	if err := sock.SetRcvtimeo(recvTimeout); err != nil {
		fail(err.Error())
	}
	if err := sock.Connect(endpoint); err != nil {
		fail(fmt.Sprintf("connect %s: %v", endpoint, err))
	}
	return sock
}

func recv3(sock *zmq.Socket, what string) (string, []byte, uint32) {
	parts, err := sock.RecvMessageBytes(0)
	if err != nil || len(parts) < 3 {
		fail(fmt.Sprintf("no %s notification within %s — is the matching zmqpub* set? (see docs/troubleshooting.md)", what, recvTimeout))
	}
	return string(parts[0]), parts[1], binary.LittleEndian.Uint32(parts[2][:4])
}

func main() {
	fmt.Println("=== Lab 06: ZMQ Listener (Go) ===")

	fmt.Print("[Step 1] Bootstrapping lab wallet & funds ... ")
	res, err := common.BootstrapLab(nil)
	if err != nil {
		fail(err.Error())
	}
	node, wallet := res.RPC, res.WalletRPC
	cfg := common.GetConfig()
	fmt.Println("✓")

	fmt.Printf("[Step 2] Subscribing to rawtx (%s) and rawblock (%s) ... ", cfg.ZMQRawTx, cfg.ZMQRawBlock)
	subTx := subscribe(cfg.ZMQRawTx, "rawtx")
	defer subTx.Close()
	subBlock := subscribe(cfg.ZMQRawBlock, "rawblock")
	defer subBlock.Close()
	time.Sleep(500 * time.Millisecond) // let the SUB subscriptions propagate
	fmt.Println("✓")

	fmt.Print("[Step 3] Broadcasting a transaction ... ")
	destRaw, _ := wallet.Call("getnewaddress", "zmq_probe", "bech32m")
	sentRaw, err := wallet.Call("sendtoaddress", str(destRaw), 0.01)
	if err != nil {
		fail(err.Error())
	}
	sentTxid := str(sentRaw)
	fmt.Printf("✓ (txid: %s…)\n", sentTxid[:16])

	fmt.Print("[Step 4] Received rawtx frame & verified txid ... ")
	topic, body, seq := recv3(subTx, "rawtx")
	if topic != "rawtx" {
		fail("unexpected topic on rawtx socket")
	}
	decRaw, err := node.Call("decoderawtransaction", hex.EncodeToString(body))
	if err != nil {
		fail(err.Error())
	}
	var dec struct {
		Txid string `json:"txid"`
	}
	if err := json.Unmarshal(decRaw, &dec); err != nil || dec.Txid != sentTxid {
		fail(fmt.Sprintf("rawtx txid %s != %s", dec.Txid, sentTxid))
	}
	fmt.Printf("✓ (seq %d)\n", seq)

	fmt.Print("[Step 5] Mining 1 block ... ")
	minerRaw, _ := wallet.Call("getnewaddress", "lab06_miner", "bech32m")
	genRaw, err := wallet.Call("generatetoaddress", 1, str(minerRaw))
	if err != nil {
		fail(err.Error())
	}
	var hashes []string
	_ = json.Unmarshal(genRaw, &hashes)
	if len(hashes) != 1 {
		fail("generatetoaddress did not return a block hash")
	}
	blockHash := hashes[0]
	fmt.Printf("✓ (hash: %s…)\n", blockHash[:16])

	fmt.Print("[Step 6] Received rawblock frame & verified against getblock ... ")
	topic, body, seq = recv3(subBlock, "rawblock")
	if topic != "rawblock" {
		fail("unexpected topic on rawblock socket")
	}
	expRaw, err := node.Call("getblock", blockHash, 0)
	if err != nil {
		fail(err.Error())
	}
	if str(expRaw) != hex.EncodeToString(body) {
		fail("rawblock payload does not match getblock")
	}
	fmt.Printf("✓ (seq %d)\n", seq)

	fmt.Println("======================================================")
	fmt.Println("Result: PASS ✓")
	fmt.Println("======================================================")
	os.Exit(0)
}
