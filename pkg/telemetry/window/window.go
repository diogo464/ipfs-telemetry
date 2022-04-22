package window

import (
	"sort"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
)

type FetchResult struct {
	NextSeqN  uint64
	Snapshots []*pb.Snapshot
}

type Stats struct {
	Count  map[string]uint32
	Memory map[string]uint32
}

type Window interface {
	PushSnapshot(snapshot.Snapshot)
	PushEvent(snapshot.Snapshot)
	// Fetch snapshots from range [since, since + n)
	// If `since` is not in the window then it is moved forward until it is
	Fetch(since uint64, n uint64) FetchResult
	FetchAll() FetchResult
	Stats(out *Stats)
}

type SnapshotSinkAdaptor struct{ w Window }

func (s SnapshotSinkAdaptor) Push(snap snapshot.Snapshot) {
	s.w.PushSnapshot(snap)
}

func SnapshotSink(w Window) snapshot.Sink {
	return SnapshotSinkAdaptor{w}
}

type EventSinkAdapator struct{ w Window }

func (s EventSinkAdapator) Push(snap snapshot.Snapshot) {
	s.w.PushEvent(snap)
}

func EventSink(w Window) snapshot.Sink {
	return EventSinkAdapator{w}
}

type windowItem struct {
	seqn      uint64
	snapshot  *pb.Snapshot
	timestamp time.Time
	size      uint32
	name      string
}

// copy [since, until)
func copySinceSeqN(v *vecdeque[windowItem], since uint64, until uint64, out []*pb.Snapshot) []*pb.Snapshot {
	start := sort.Search(v.Len(), func(i int) bool {
		return v.Get(i).seqn >= since
	})
	for i := start; i < v.Len(); i++ {
		item := v.Get(i)
		if item.seqn >= until {
			break
		}
		out = append(out, item.snapshot)
	}
	return out
}
