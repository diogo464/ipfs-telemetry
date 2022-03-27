package telemetry

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/telemetry/utils"
	"git.d464.sh/adc/telemetry/telemetry/wire"
	"github.com/google/uuid"
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
	s uuid.UUID
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
		s: uuid.New(),
		n: n,
		o: o,
		w: snapshot.NewWindow(o.windowDuration),
	}

	h := n.PeerHost
	h.SetStreamHandler(ID, t.TelemetryHandler)

	go snapshot.NewPingCollector(t.host(), t, snapshot.PingOptions{
		PingCount: 5,
		Interval:  time.Second * 5,
		Timeout:   time.Second * 10,
	}).Run()

	go snapshot.NewRoutingTableCollector(t.n, t, snapshot.RoutingTableOptions{
		Interval: time.Second * 10,
	}).Run()

	go snapshot.NewNetworkCollector(t.n, t, snapshot.NetworkOptions{
		Interval: time.Second * 15,
	}).Run()

	return t, nil
}

func (t *TelemetryService) Push(snapshot *snapshot.Snapshot) {
	t.l.Lock()
	defer t.l.Unlock()
	t.w.Push(snapshot)
}

func (t *TelemetryService) TelemetryHandler(s network.Stream) {
	defer s.Close()
	var err error = nil

	request, err := wire.ReadRequest(context.TODO(), s)
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
	log.Println("Handle Request Snapshot")

	var since uint64 = 0
	if r.Session == t.s {
		since = r.Since
	}

	t.l.Lock()
	snapshots := t.w.Since(since)
	nextseqn := t.w.NextSeqN()
	t.l.Unlock()

	fmt.Println("Request")
	fmt.Println("  Since =", since)
	fmt.Println("  Next =", nextseqn)

	index := 0
	for index < len(snapshots) {
		if !r.MatchTag(snapshots[index].Tag) {
			fmt.Println("Ignoring snapshot", snapshots[index].Tag)
			snapshots[index] = snapshots[len(snapshots)-1]
			snapshots = snapshots[:len(snapshots)-1]
		} else {
			index += 1
		}
	}

	response := wire.NewResponseSnapshot(t.s, snapshots, nextseqn)
	if err := wire.WriteResponse(context.TODO(), s, response); err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleRequestSystemInfo(s network.Stream) {
	log.Println("Handle Request SystemInfo")
	response := wire.NewResponseSystemInfo()
	if err := wire.WriteResponse(context.TODO(), s, response); err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleBandwidthDownload(s network.Stream) {
	log.Println("Handle BandwidthDownload")
	io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
}

func (t *TelemetryService) handleBandwidthUpload(s network.Stream) {
	log.Println("Handle BandwidthUpload")
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
