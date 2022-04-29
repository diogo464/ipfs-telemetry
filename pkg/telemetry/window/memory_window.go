package window

import (
	"math"
	"sync"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/utils"
)

var _ Window = (*MemoryWindow)(nil)

type MemoryWindow struct {
	mu        sync.Mutex
	duration  time.Duration
	maxEvents int

	nextSeqN   uint64
	datapoints *vecdeque[windowItem]
	events     *vecdeque[windowItem]
	stats      *Stats
}

func NewMemoryWindow(duration time.Duration, maxEvents int) *MemoryWindow {
	return &MemoryWindow{
		mu:        sync.Mutex{},
		duration:  duration,
		maxEvents: maxEvents,

		nextSeqN:   0,
		datapoints: newVecDeque[windowItem](),
		events:     newVecDeque[windowItem](),
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

	datapoints := make([]*pb.Datapoint, 0)
	until := since + n
	datapoints = copySinceSeqN(w.datapoints, since, until, datapoints)
	datapoints = copySinceSeqN(w.events, since, until, datapoints)

	return FetchResult{
		NextSeqN:   utils.Min(until, w.nextSeqN),
		Datapoints: datapoints,
	}
}

// FetchAll implements Window
func (w *MemoryWindow) FetchAll() FetchResult {
	return w.Fetch(0, math.MaxUint64)
}

// PushEvent implements Window
func (w *MemoryWindow) PushEvent(s datapoint.Datapoint) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pushToVec(s, w.events)
}

// PushSnapshot implements Window
func (w *MemoryWindow) PushSnapshot(s datapoint.Datapoint) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pushToVec(s, w.datapoints)
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

func (w *MemoryWindow) pushToVec(s datapoint.Datapoint, v *vecdeque[windowItem]) {
	w.clean()

	name := s.GetName()
	size := s.GetSizeEstimate()
	w.stats.Count[name] += 1
	w.stats.Memory[name] += size

	seqn := w.nextSeqN
	w.nextSeqN += 1
	v.PushBack(windowItem{
		seqn:        seqn,
		datapointpb: s.ToPB(),
		timestamp:   time.Now(),
		size:        size,
		name:        name,
	})
}

func (w *MemoryWindow) clean() {
	for (!w.events.IsEmpty() && time.Since(w.events.Front().timestamp) > w.duration) || (w.events.Len() > w.maxEvents) {
		item := w.events.PopFront()
		w.stats.Count[item.name] -= 1
		w.stats.Memory[item.name] -= item.size
	}

	for !w.datapoints.IsEmpty() && time.Since(w.datapoints.Front().timestamp) > w.duration {
		item := w.datapoints.PopFront()
		w.stats.Count[item.name] -= 1
		w.stats.Memory[item.name] -= item.size
	}
}
