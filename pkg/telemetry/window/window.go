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
	items    *vecdeque[windowItem]
	duration time.Duration
}

func NewWindow(duration time.Duration) Window {
	return newWindowImpl(duration)
}

func newWindowImpl(duration time.Duration) *windowImpl {
	return &windowImpl{
		seqn:     1,
		items:    newVecDeque[windowItem](),
		duration: duration,
	}
}

func (w *windowImpl) push(t time.Time, v *pb.Snapshot) {
	w.Lock()
	defer w.Unlock()

	w.clean()
	seqn := w.seqn
	w.seqn += 1
	w.items.PushBack(windowItem{
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

	if w.items.IsEmpty() {
		return nil
	}

	left := w.items.Front().seqn
	if seqn < left {
		seqn = left
	}

	start := int(seqn - left)
	size := w.items.Len() - start
	if size <= 0 {
		return nil
	}

	snapshots := make([]*pb.Snapshot, 0, size)
	for i := start; i < w.items.Len(); i++ {
		snapshots = append(snapshots, w.items.Get(i).snapshot)
	}

	return snapshots
}

func (w *windowImpl) NextSeqN() uint64 {
	w.Lock()
	defer w.Unlock()

	return w.seqn
}

func (w *windowImpl) clean() {
	for !w.items.IsEmpty() && time.Since(w.items.Front().timestamp) > w.duration {
		w.items.PopFront()
	}
}
