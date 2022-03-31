package window

import (
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"git.d464.sh/adc/telemetry/plugin/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Window interface {
	snapshot.Sink
	Since(seqn uint64) *pb.Snapshot_Set
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

func (w *windowImpl) Since(seqn uint64) *pb.Snapshot_Set {
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

	pings := make([]*pb.Snapshot_Ping, 0)
	routingtables := make([]*pb.Snapshot_RoutingTable, 0)
	networks := make([]*pb.Snapshot_Network, 0)
	for i := start; i < len(w.items); i++ {
		switch v := w.items[i].snapshot.(type) {
		case *pb.Snapshot_Ping:
			pings = append(pings, v)
		case *pb.Snapshot_RoutingTable:
			routingtables = append(routingtables, v)
		case *pb.Snapshot_Network:
			networks = append(networks, v)
		default:
		}
	}

	return &pb.Snapshot_Set{
		Pings:         pings,
		RoutingTables: routingtables,
		Networks:      networks,
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
