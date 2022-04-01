package window

import (
	"sync"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Window = (*windowImpl)(nil)

type Window interface {
	snapshot.Sink
	Since(seqn uint64) *pb.Set
	NextSeqN() uint64
}

type windowItem struct {
	seqn      uint64
	snapshot  interface{}
	timestamp time.Time
}

type windowImpl struct {
	sync.Mutex
	seqn     uint64
	items    []windowItem
	duration time.Duration
}

func NewWindow(duration time.Duration) Window {
	return newWindowImpl(duration)
}

func newWindowImpl(duration time.Duration) *windowImpl {
	return &windowImpl{
		seqn:     1,
		items:    []windowItem{},
		duration: duration,
	}
}

func (w *windowImpl) push(t time.Time, v interface{}) {
	w.clean()
	seqn := w.seqn
	w.seqn += 1
	w.items = append(w.items, windowItem{
		seqn:      seqn,
		snapshot:  v,
		timestamp: t,
	})
}

func (w *windowImpl) PushPing(ping *snapshot.Ping) {
	w.push(ping.Timestamp, ping.ToPB())
}

func (w *windowImpl) PushRoutingTable(rt *snapshot.RoutingTable) {
	w.push(rt.Timestamp, rt.ToPB())
}

func (w *windowImpl) PushNetwork(n *snapshot.Network) {
	w.push(n.Timestamp, n.ToPB())
}

func (w *windowImpl) PushResources(r *snapshot.Resources) {
	w.push(r.Timestamp, r.ToPB())
}

func (w *windowImpl) PushBitswap(b *snapshot.Bitswap) {
	w.push(b.Timestamp, b.ToPB())
}

func (w *windowImpl) Since(seqn uint64) *pb.Set {
	w.Lock()
	defer w.Unlock()

	if len(w.items) == 0 {
		return nil
	}

	left := w.items[0].seqn
	if seqn < left {
		seqn = left
	}

	start := int(seqn - left)
	size := len(w.items) - start
	if size <= 0 {
		return nil
	}

	pings := make([]*pb.Ping, 0)
	routingtables := make([]*pb.RoutingTable, 0)
	networks := make([]*pb.Network, 0)
	resources := make([]*pb.Resources, 0)
	bitswap := make([]*pb.Bitswap, 0)
	for i := start; i < len(w.items); i++ {
		switch v := w.items[i].snapshot.(type) {
		case *pb.Ping:
			pings = append(pings, v)
		case *pb.RoutingTable:
			routingtables = append(routingtables, v)
		case *pb.Network:
			networks = append(networks, v)
		case *pb.Resources:
			resources = append(resources, v)
		case *pb.Bitswap:
			bitswap = append(bitswap, v)
		default:
			// TODO: remove this
			panic("unimplemented")
		}
	}

	return &pb.Set{
		Pings:         pings,
		RoutingTables: routingtables,
		Networks:      networks,
		Resources:     resources,
		Bitswaps:      bitswap,
	}
}

func (w *windowImpl) NextSeqN() uint64 {
	w.Lock()
	defer w.Unlock()

	return w.seqn
}

func (w *windowImpl) clean() {
	for len(w.items) > 0 && time.Since(w.items[0].timestamp) > w.duration {
		w.items = w.items[1:]
	}
}

func unpackPeerAddrInfo(info *peer.AddrInfo) (string, []string) {
	pid := info.ID.Pretty()
	addrs := make([]string, 0, len(info.Addrs))
	for _, addr := range info.Addrs {
		addrs = append(addrs, addr.String())
	}
	return pid, addrs
}
