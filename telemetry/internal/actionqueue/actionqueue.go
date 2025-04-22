package actionqueue

import (
	"container/heap"
	"math"
	"time"
)

type Action[T any] struct {
	Data     *T
	Deadline time.Time
}

type actionHeap[T any] []Action[T]

func (h actionHeap[T]) Len() int           { return len(h) }
func (h actionHeap[T]) Less(i, j int) bool { return h[i].Deadline.Before(h[j].Deadline) }
func (h actionHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (pq *actionHeap[T]) Push(x any) {
	item := x.(Action[T])
	*pq = append(*pq, item)
}
func (pq *actionHeap[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1].Data = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

type Queue[T any] struct {
	actions actionHeap[T]
}

func Now[T any](data *T) Action[T] {
	return Action[T]{
		Data:     data,
		Deadline: time.Now(),
	}
}

func After[T any](data *T, dur time.Duration) Action[T] {
	return Action[T]{
		Data:     data,
		Deadline: time.Now().Add(dur),
	}
}

func Deadline[T any](data *T, deadline time.Time) Action[T] {
	return Action[T]{
		Data:     data,
		Deadline: deadline,
	}
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		actions: make([]Action[T], 0),
	}
}

func (w *Queue[T]) Push(action Action[T]) {
	heap.Push(&w.actions, action)
}

func (w *Queue[T]) Pop() *T {
	action := heap.Pop(&w.actions).(Action[T])
	return action.Data
}

func (w *Queue[T]) Len() int {
	return len(w.actions)
}

func (w *Queue[T]) TimeUntilAction() time.Duration {
	if len(w.actions) == 0 {
		return time.Duration(math.MaxInt64)
	} else {
		return time.Until(w.actions[0].Deadline)
	}
}

func (w *Queue[T]) TimerUntilAction() *time.Timer {
	return time.NewTimer(w.TimeUntilAction())
}
