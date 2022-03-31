module git.d464.sh/adc/telemetry/plugin

go 1.16

require (
	github.com/google/uuid v1.3.0
	github.com/ipfs/go-ipfs v0.12.1
	github.com/libp2p/go-libp2p v0.16.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.11.0
	github.com/libp2p/go-libp2p-gostream v0.3.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/ipfs/go-ipfs => ../third_party/go-ipfs

replace github.com/libp2p/go-libp2p-kad-dht v0.15.0 => ../third_party/go-libp2p-kad-dht/

replace github.com/libp2p/go-libp2p-kbucket v0.4.7 => ../third_party/go-libp2p-kbucket/
