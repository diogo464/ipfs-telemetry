module git.d464.sh/adc/telemetry

go 1.16

require (
	github.com/friendsofgo/errors v0.9.2
	github.com/google/uuid v1.3.0
	github.com/ipfs/go-bitswap v0.6.0
	github.com/ipfs/go-datastore v0.5.1
	github.com/ipfs/go-ipfs v0.12.1
	github.com/kat-co/vala v0.0.0-20170210184112-42e1d8b61f12
	github.com/lib/pq v1.10.4
	github.com/libp2p/go-libp2p v0.16.0
	github.com/libp2p/go-libp2p-connmgr v0.2.4
	github.com/libp2p/go-libp2p-core v0.11.0
	github.com/libp2p/go-libp2p-gostream v0.3.0
	github.com/libp2p/go-libp2p-kad-dht v0.15.0
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/prometheus/client_golang v1.11.0
	github.com/shirou/gopsutil v2.21.11+incompatible
	github.com/spf13/viper v1.9.0
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/urfave/cli/v2 v2.0.0
	github.com/volatiletech/randomize v0.0.1
	github.com/volatiletech/sqlboiler/v4 v4.8.6
	github.com/volatiletech/strmangle v0.0.2
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/atomic v1.9.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/ipfs/go-ipfs => ./third_party/go-ipfs/

replace github.com/ipfs/go-bitswap => ./third_party/go-bitswap/

replace github.com/libp2p/go-libp2p-kad-dht => ./third_party/go-libp2p-kad-dht/

replace github.com/libp2p/go-libp2p-kbucket => ./third_party/go-libp2p-kbucket/

replace git.d464.sh/adc/telemetry => ./
