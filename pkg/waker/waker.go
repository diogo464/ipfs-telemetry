package waker

import (
	"container/heap"
	"context"
	"math"
	"time"
)

type wakerItem[T any] struct {
	deadline time.Time
	data     *T
}

type wakerItemHeap[T any] []wakerItem[T]

func (h wakerItemHeap[T]) Len() int           { return len(h) }
func (h wakerItemHeap[T]) Less(i, j int) bool { return h[i].deadline.Before(h[j].deadline) }
func (h wakerItemHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (pq *wakerItemHeap[T]) Push(x any) {
	item := x.(wakerItem[T])
	*pq = append(*pq, item)
}
func (pq *wakerItemHeap[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1].data = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

type Waker[T any] struct {
	cpush    chan wakerItem[T]
	creceive chan *T

	// task data
	heap   wakerItemHeap[T]
	cancel context.CancelFunc
}

func NewWaker[T any]() *Waker[T] {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Waker[T]{
		cpush:    make(chan wakerItem[T]),
		creceive: make(chan *T),
		heap:     wakerItemHeap[T]([]wakerItem[T]{}),
		cancel:   cancel,
	}
	go w.wakerTask(ctx)
	return w
}

func (w *Waker[T]) Push(v *T, delay time.Duration) {
	w.PushDeadline(v, time.Now().Add(delay))
}

func (w *Waker[T]) PushNow(v *T) {
	w.PushDeadline(v, time.Now())
}

func (w *Waker[T]) PushDeadline(v *T, deadline time.Time) {
	w.cpush <- wakerItem[T]{
		deadline: deadline,
		data:     v,
	}
}

func (w *Waker[T]) Receive() <-chan *T {
	return w.creceive
}

func (w *Waker[T]) Close() {
	w.cancel()
}

func (w *Waker[T]) wakerTask(ctx context.Context) {
	timer := time.NewTimer(time.Duration(math.MaxInt64))
	for {
		select {
		case item := <-w.cpush:
			heap.Push(&w.heap, item)
			timerDelay := time.Until(w.heap[0].deadline)
			timer.Reset(timerDelay)
		case <-timer.C:
			item := heap.Pop(&w.heap).(wakerItem[T])
			w.creceive <- item.data
		case <-ctx.Done():
			return
		}
	}
}
