package window

import (
	"sync"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
)

var _ Window = (*windowImpl)(nil)

type Window interface {
	snapshot.Sink
	Since(seqn uint64) []*pb.Snapshot
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

func (w *windowImpl) push(t time.Time, v *pb.Snapshot) {
	w.Lock()
	defer w.Unlock()

	w.clean()
	seqn := w.seqn
	w.seqn += 1
	w.items = append(w.items, windowItem{
		seqn:      seqn,
		snapshot:  v,
		timestamp: t,
	})
}

func (w *windowImpl) Push(s snapshot.Snapshot) {
	w.push(s.GetTimestamp(), s.ToPB())
}

func (w *windowImpl) Since(seqn uint64) []*pb.Snapshot {
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

	snapshots := make([]*pb.Snapshot, 0, size)
	for i := start; i < len(w.items); i++ {
		snapshots = append(snapshots, w.items[i].snapshot)
	}

	return snapshots
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
