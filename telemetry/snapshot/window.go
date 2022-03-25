package snapshot

import (
	"time"
)

type Window interface {
	Push(snapshot *Snapshot)
	Since(seqn uint64) []*Snapshot
	NextSeqN() uint64
}

type windowItem struct {
	seqn      uint64
	snapshot  *Snapshot
	timestamp time.Time
}

type windowImpl struct {
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

func (w *windowImpl) Push(snapshot *Snapshot) {
	w.clean()
	seqn := w.seqn
	w.seqn += 1
	w.items = append(w.items, windowItem{
		seqn:      seqn,
		snapshot:  snapshot,
		timestamp: time.Now(),
	})
}

func (w *windowImpl) Since(seqn uint64) []*Snapshot {
	if len(w.items) == 0 {
		return []*Snapshot{}
	}

	left := w.items[0].seqn
	if seqn < left {
		seqn = left
	}

	start := int(seqn - left)
	size := len(w.items) - start
	if size <= 0 {
		return []*Snapshot{}
	}

	snapshots := make([]*Snapshot, 0, size)
	for i := start; i < len(w.items); i++ {
		snapshots = append(snapshots, w.items[i].snapshot)
	}

	return snapshots
}

func (w *windowImpl) NextSeqN() uint64 {
	return w.seqn
}

func (w *windowImpl) clean() {
	for len(w.items) > 0 && time.Since(w.items[0].timestamp) > w.duration {
		w.items = w.items[1:]
	}
}
