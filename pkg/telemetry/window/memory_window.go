package window

import (
	"math"
	"sync"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/pkg/utils"
)

var _ Window = (*MemoryWindow)(nil)

type MemoryWindow struct {
	mu        sync.Mutex
	duration  time.Duration
	maxEvents int

	nextSeqN  uint64
	snapshots *vecdeque[windowItem]
	events    *vecdeque[windowItem]
	stats     *Stats
}

func NewMemoryWindow(duration time.Duration, maxEvents int) *MemoryWindow {
	return &MemoryWindow{
		mu:        sync.Mutex{},
		duration:  duration,
		maxEvents: maxEvents,

		nextSeqN:  0,
		snapshots: newVecDeque[windowItem](),
		events:    newVecDeque[windowItem](),
		stats: &Stats{
			Count:  make(map[string]uint32),
			Memory: make(map[string]uint32),
		},
	}
}

// Fetch implements Window
func (w *MemoryWindow) Fetch(since uint64, n uint64) FetchResult {
	w.mu.Lock()
	defer w.mu.Unlock()

	datapoints := make([]*pb.Snapshot, 0)
	until := since + n
	datapoints = copySinceSeqN(w.snapshots, since, until, datapoints)
	datapoints = copySinceSeqN(w.events, since, until, datapoints)

	return FetchResult{
		NextSeqN:  utils.Min(until, w.nextSeqN),
		Snapshots: datapoints,
	}
}

// FetchAll implements Window
func (w *MemoryWindow) FetchAll() FetchResult {
	return w.Fetch(0, math.MaxUint64)
}

// PushEvent implements Window
func (w *MemoryWindow) PushEvent(s snapshot.Snapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pushToVec(s, w.events)
}

// PushSnapshot implements Window
func (w *MemoryWindow) PushSnapshot(s snapshot.Snapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pushToVec(s, w.snapshots)
}

// Stats implements Window
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

func (w *MemoryWindow) pushToVec(s snapshot.Snapshot, v *vecdeque[windowItem]) {
	w.clean()

	name := s.GetName()
	size := s.GetSizeEstimate()
	w.stats.Count[name] += 1
	w.stats.Memory[name] += size

	seqn := w.nextSeqN
	w.nextSeqN += 1
	v.PushBack(windowItem{
		seqn:      seqn,
		snapshot:  s.ToPB(),
		timestamp: time.Now(),
		size:      size,
		name:      name,
	})
}

func (w *MemoryWindow) clean() {
	for (!w.events.IsEmpty() && time.Since(w.events.Front().timestamp) > w.duration) || (w.events.Len() > w.maxEvents) {
		item := w.events.PopFront()
		w.stats.Count[item.name] -= 1
		w.stats.Memory[item.name] -= item.size
	}

	for !w.snapshots.IsEmpty() && time.Since(w.snapshots.Front().timestamp) > w.duration {
		item := w.snapshots.PopFront()
		w.stats.Count[item.name] -= 1
		w.stats.Memory[item.name] -= item.size
	}
}
