package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/plugin/snapshot"
	"git.d464.sh/adc/telemetry/plugin/wire"
	"github.com/google/uuid"
	"github.com/ipfs/go-ipfs/core"
)

const (
	ID                     = "/telemetry/telemetry/0.0.0"
	BANDWIDTH_PAYLOAD_SIZE = 32 * 1024 * 1024
)

type TelemetryService struct {
	// current session, randomly generated number
	s uuid.UUID
	// the node we are collecting telemetry from
	n *core.IpfsNode
	// read-only options
	o *options
	w wire.Window
}

func NewTelemetryService(n *core.IpfsNode, opts ...Option) (*TelemetryService, error) {
	o := new(options)
	defaults(o)

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	t := &TelemetryService{
		s: uuid.New(),
		n: n,
		o: o,
		w: wire.NewWindow(o.windowDuration),
	}

	h := n.PeerHost
	h.SetStreamHandler(ID, t.TelemetryHandler)

	go snapshot.NewPingCollector(t.n.PeerHost, t.w, snapshot.PingOptions{
		PingCount: 5,
		Interval:  time.Second * 5,
		Timeout:   time.Second * 10,
	}).Run()

	go snapshot.NewRoutingTableCollector(t.n, t.w, snapshot.RoutingTableOptions{
		Interval: time.Second * 10,
	}).Run()

	go snapshot.NewNetworkCollector(t.n, t.w, snapshot.NetworkOptions{
		Interval: time.Second * 15,
	}).Run()

	return t, nil
}
