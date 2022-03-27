module git.d464.sh/adc/telemetry/monitor

go 1.16

require (
	git.d464.sh/adc/telemetry/telemetry v0.0.0-00010101000000-000000000000
	github.com/friendsofgo/errors v0.9.2
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/uuid v1.3.0
	github.com/ipfs/go-datastore v0.5.1
	github.com/kat-co/vala v0.0.0-20170210184112-42e1d8b61f12
	github.com/lib/pq v1.10.4
	github.com/libp2p/go-libp2p v0.18.0
	github.com/libp2p/go-libp2p-core v0.14.0
	github.com/libp2p/go-libp2p-kad-dht v0.15.0
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.9.0
	github.com/urfave/cli/v2 v2.4.0
	github.com/volatiletech/randomize v0.0.1
	github.com/volatiletech/sqlboiler/v4 v4.8.6
	github.com/volatiletech/strmangle v0.0.2
)

replace git.d464.sh/adc/telemetry/telemetry => ../telemetry/
