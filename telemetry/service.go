package telemetry

import (
	"fmt"
	"io"
	"math/rand"
	"sync"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/telemetry/utils"
	"git.d464.sh/adc/telemetry/telemetry/wire"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
)

const (
	ID                     = "/telemetry/telemetry/0.0.0"
	BANDWIDTH_PAYLOAD_SIZE = 32 * 1024 * 1024
)

type TelemetryService struct {
	// current session, randomly generated number
	s uint64
	// the node we are collecting telemetry from
	n *core.IpfsNode
	// read-only options
	o *options
	l sync.Mutex
	w snapshot.Window
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
		s: uint64(rand.Int63()),
		n: n,
		o: o,
		w: snapshot.NewWindow(o.windowDuration),
	}

	h := n.PeerHost
	h.SetStreamHandler(ID, t.TelemetryHandler)

	go t.collectorPing()
	go t.collectorRT()
	go t.collectorNetwork()

	return t, nil
}

func (t *TelemetryService) TelemetryHandler(s network.Stream) {
	defer s.Close()
	var err error = nil

	request, err := wire.ReadRequest(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(request)

	switch request.Type {
	case wire.REQUEST_SNAPSHOT:
		t.handleRequestSnapshot(s, request.GetSince())
	case wire.REQUEST_SYSTEM_INFO:
		t.handleRequestSystemInfo(s)
	case wire.REQUEST_BANDWDITH_DOWNLOAD:
		t.handleBandwidthDownload(s)
	case wire.REQUEST_BANDWDITH_UPLOAD:
		t.handleBandwidthUpload(s)
	default:
		panic("unreachable")
	}

	if err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleRequestSnapshot(s network.Stream, r *wire.RequestSnapshot) {
	t.l.Lock()
	snapshots := t.w.Since(r.Since)
	t.l.Unlock()

	response := wire.NewResponseSnapshot(t.s, snapshots)
	if err := wire.WriteResponse(s, response); err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleRequestSystemInfo(s network.Stream) {
	response := wire.NewResponseSystemInfo()
	if err := wire.WriteResponse(s, response); err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleBandwidthDownload(s network.Stream) {
	io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
}

func (t *TelemetryService) handleBandwidthUpload(s network.Stream) {
	io.Copy(s, io.LimitReader(utils.NullReader{}, BANDWIDTH_PAYLOAD_SIZE))
}

func (t *TelemetryService) host() host.Host {
	return t.n.PeerHost
}

func (t *TelemetryService) pushSnapshot(snapshot *snapshot.Snapshot) {
	t.l.Lock()
	defer t.l.Unlock()
	t.w.Push(snapshot)
}
