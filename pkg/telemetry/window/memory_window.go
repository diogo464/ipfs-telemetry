package window

import (
	"math"
	"sort"
	"sync"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"git.d464.sh/adc/telemetry/pkg/utils"
)

var _ Window = (*MemoryWindow)(nil)

type MemoryWindow struct {
	mu       sync.Mutex
	duration time.Duration
	items    *vecdeque[windowItem]
	stats    *Stats
}

func NewMemoryWindow(duration time.Duration) *MemoryWindow {
	return &MemoryWindow{
		mu:       sync.Mutex{},
		duration: duration,
		items:    newVecDeque[windowItem](),
		stats: &Stats{
			SnapshotCounts: map[string]int{},
			MemoryCounts:   map[string]int{},
		},
	}
}

// Push implements Window
func (w *MemoryWindow) Push(s snapshot.Snapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	name := s.GetName()
	size := snapshot.SnapshotSize(s)
	w.stats.SnapshotCounts[name] += 1
	w.stats.MemoryCounts[name] += size

	w.items.PushBack(windowItem{
		seqn:      nextSeqN(w.items),
		snapshot:  s.ToPB(),
		timestamp: s.GetTimestamp(),
		size:      size,
		name:      name,
	})
	w.clean()
}

// Fetch implements Window
func (w *MemoryWindow) Fetch(since uint64, n int) FetchResult {
	w.mu.Lock()
	defer w.mu.Unlock()

	if n == 0 || w.items.IsEmpty() || since > w.items.Back().seqn {
		return FetchResult{
			FirstSeqN: nextSeqN(w.items),
			Snapshots: []*pb.Snapshot{},
		}
	}

	start := sort.Search(w.items.Len(), func(i int) bool {
		return w.items.Get(i).seqn >= since
	})
	end := utils.Min(start+n, w.items.Len())
	snapshots := make([]*pb.Snapshot, 0, end-start)
	for i := start; i < end; i++ {
		snapshots = append(snapshots, w.items.Get(i).snapshot)
	}
	return FetchResult{
		FirstSeqN: w.items.Get(start).seqn,
		Snapshots: snapshots,
	}
}

func (w *MemoryWindow) FetchAll() FetchResult {
	return w.Fetch(0, math.MaxInt)
}

func (w *MemoryWindow) Stats(out *Stats) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if out.SnapshotCounts == nil {
		out.SnapshotCounts = make(map[string]int)
	}
	if out.MemoryCounts == nil {
		out.MemoryCounts = make(map[string]int)
	}

	for k, v := range w.stats.SnapshotCounts {
		out.SnapshotCounts[k] = v
	}
	for k, v := range w.stats.MemoryCounts {
		out.MemoryCounts[k] = v
	}
}

func (w *MemoryWindow) clean() {
	for !w.items.IsEmpty() && time.Since(w.items.Front().timestamp) > w.duration {
		item := w.items.PopFront()
		w.stats.SnapshotCounts[item.name] -= 1
		w.stats.MemoryCounts[item.name] -= item.size
	}
}
