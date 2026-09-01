module github.com/Akpahsamuel/Bitcoin-simulator/examples/02-keys-and-addresses/go

go 1.22

require (
	github.com/Akpahsamuel/Bitcoin-simulator/examples/go v0.0.0
	github.com/btcsuite/btcd v0.24.2
	github.com/btcsuite/btcd/btcec/v2 v2.3.4
	github.com/btcsuite/btcd/btcutil v1.1.5
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0
)

replace github.com/Akpahsamuel/Bitcoin-simulator/examples/go => ../../go
