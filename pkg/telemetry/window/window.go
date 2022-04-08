package window

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
)

type Stats struct {
	SnapshotCounts map[string]int
	MemoryCounts   map[string]int
}

type Window interface {
	snapshot.Sink
	// Fetch snapshots from range [since, since + n)
	// If `since` is not in the window then it is moved forward until it is
	Fetch(since uint64, n int) FetchResult
	FetchAll() FetchResult
	Stats(out *Stats)
}

type FetchResult struct {
	FirstSeqN uint64
	Snapshots []*pb.Snapshot
}

type windowItem struct {
	seqn      uint64
	snapshot  *pb.Snapshot
	timestamp time.Time
	size      int
	name      string
}

func nextSeqN(v *vecdeque[windowItem]) uint64 {
	if v.IsEmpty() {
		return 0
	}
	return v.Back().seqn + 1
}
