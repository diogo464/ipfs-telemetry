package wire

import (
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"git.d464.sh/adc/telemetry/plugin/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Window interface {
	snapshot.Sink
	Since(seqn uint64) *pb.Response_Snapshots
	NextSeqN() uint64
}

type windowItem struct {
	seqn      uint64
	snapshot  *pb.Snapshot
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

func (w *windowImpl) push(snapshot *pb.Snapshot) {
	w.clean()
	seqn := w.seqn
	w.seqn += 1
	w.items = append(w.items, windowItem{
		seqn:      seqn,
		snapshot:  snapshot,
		timestamp: time.Now(),
	})
}

func (w *windowImpl) PushPing(ping *snapshot.Ping) {
	source_pid, source_addrs := unpackPeerAddrInfo(&ping.Source)
	dest_pid, dest_addrs := unpackPeerAddrInfo(&ping.Destination)
	durations := make([]*durationpb.Duration, 0, len(ping.Durations))
	for _, dur := range ping.Durations {
		durations = append(durations, durationpb.New(dur))
	}

	w.push(&pb.Snapshot{
		Body: &pb.Snapshot_Ping_{Ping: &pb.Snapshot_Ping{
			SourcePid:        source_pid,
			SourceAddrs:      source_addrs,
			DestinationPid:   dest_pid,
			DestinationAddrs: dest_addrs,
			Durations:        []*durationpb.Duration{},
		}},
	})
}

func (w *windowImpl) PushRoutingTable(rt *snapshot.RoutingTable) {
	w.push(&pb.Snapshot{
		Body: &pb.Snapshot_RoutingTable_{
			RoutingTable: &pb.Snapshot_RoutingTable{
				Buckets: rt.Buckets,
			},
		},
	})
}

func (w *windowImpl) PushNetwork(n *snapshot.Network) {
	w.push(&pb.Snapshot{
		Body: &pb.Snapshot_Network_{
			Network: &pb.Snapshot_Network{
				TotalIn:   n.TotalIn,
				TotalOut:  n.TotalOut,
				RateIn:    n.RateIn,
				RateOut:   n.RateOut,
				NumConns:  n.NumConns,
				LowWater:  n.LowWater,
				HighWater: n.HighWater,
			},
		},
	})
}

func (w *windowImpl) Since(seqn uint64) *pb.Response_Snapshots {
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

	snapshots := make([]*pb.Snapshot, 0)
	for i := start; i < len(w.items); i++ {
		snapshots = append(snapshots, w.items[i].snapshot)
	}

	return &pb.Response_Snapshots{
		Next:      w.seqn,
		Snapshots: snapshots,
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
