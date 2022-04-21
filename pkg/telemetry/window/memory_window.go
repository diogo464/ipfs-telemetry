package window

import (
	"math"
	"sort"
	"sync"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/pkg/utils"
)

var _ Window = (*MemoryWindow)(nil)

type MemoryWindow struct {
	mu       sync.Mutex
	duration time.Duration
	items    *vecdeque[windowItem]
	stats    *Stats
	max      int
}

func NewMemoryWindow(duration time.Duration) *MemoryWindow {
	return NewMemoryWindowWithMax(duration, math.MaxInt)
}

func NewMemoryWindowWithMax(duration time.Duration, max int) *MemoryWindow {
	return &MemoryWindow{
		mu:       sync.Mutex{},
		duration: duration,
		items:    newVecDeque[windowItem](),
		stats: &Stats{
			Count:  map[string]uint32{},
			Memory: map[string]uint32{},
		},
		max: max,
	}
}

// Push implements Window
func (w *MemoryWindow) Push(s snapshot.Snapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	name := s.GetName()
	size := s.GetSizeEstimate()
	w.stats.Count[name] += 1
	w.stats.Memory[name] += size

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

	if out.Count == nil {
		out.Count = make(map[string]uint32)
	}
	if out.Memory == nil {
		out.Memory = make(map[string]uint32)
	}

	for k, v := range w.stats.Count {
		out.Count[k] = v
	}
	for k, v := range w.stats.Memory {
		out.Memory[k] = v
	}
}

func (w *MemoryWindow) clean() {
	for (!w.items.IsEmpty() && time.Since(w.items.Front().timestamp) > w.duration) || (w.items.Len() > w.max) {
		item := w.items.PopFront()
		w.stats.Count[item.name] -= 1
		w.stats.Memory[item.name] -= item.size
	}
}
