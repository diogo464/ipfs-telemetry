module git.d464.sh/adc/telemetry/plugin

go 1.16

require (
	git.d464.sh/adc/rle v0.0.0-20220329100220-cbaf095d04c2
	github.com/google/uuid v1.3.0
	github.com/ipfs/go-ipfs v0.12.1
	github.com/libp2p/go-libp2p v0.16.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.11.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/ipfs/go-ipfs => /var/home/diogo464/uni/adc/telemetry/plugin/../../ipfs

replace git.d464.sh/adc/rle => /var/home/diogo464/uni/adc/telemetry/plugin/../../rle
