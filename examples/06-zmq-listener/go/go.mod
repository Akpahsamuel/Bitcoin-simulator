module github.com/Akpahsamuel/Bitcoin-simulator/examples/06-zmq-listener/go

go 1.22

require (
	github.com/Akpahsamuel/Bitcoin-simulator/examples/go v0.0.0
	github.com/pebbe/zmq4 v1.2.11
)

require github.com/joho/godotenv v1.5.1 // indirect

replace github.com/Akpahsamuel/Bitcoin-simulator/examples/go => ../../go
