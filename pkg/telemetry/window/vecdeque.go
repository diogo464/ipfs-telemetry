package window

import "github.com/diogo464/telemetry/pkg/utils"

type vecdeque[T any] struct {
	buf   []T
	size  int
	start int
}

func newVecDeque[T any]() *vecdeque[T] {
	return &vecdeque[T]{
		buf:   make([]T, 0),
		size:  0,
		start: 0,
	}
}

func (v *vecdeque[T]) PushBack(x T) {
	if v.size == v.capacity() {
		v.reallocate(utils.Max(1, v.size*2))
	}
	v.buf[(v.start+v.size)%v.capacity()] = x
	v.size += 1
}

func (v *vecdeque[T]) PopFront() T {
	if v.IsEmpty() {
		panic("PopFront on empty")
	}
	var def T
	var ret T = v.buf[v.start]
	v.buf[v.start] = def
	v.start = (v.start + 1) % v.capacity()
	v.size -= 1
	return ret
}

func (v *vecdeque[T]) Get(i int) T {
	return v.buf[(v.start+i)%v.capacity()]
}

func (v *vecdeque[T]) Back() T {
	if v.IsEmpty() {
		panic("Back on empty")
	}
	return v.buf[(v.start+v.size-1)%v.capacity()]
}

func (v *vecdeque[T]) Front() T {
	if v.IsEmpty() {
		panic("Front on empty")
	}
	return v.buf[v.start]
}

func (v *vecdeque[T]) Len() int {
	return v.size
}

func (v *vecdeque[T]) IsEmpty() bool {
	return v.Len() == 0
}

func (v *vecdeque[T]) capacity() int {
	return cap(v.buf)
}

func (v *vecdeque[T]) reallocate(new_size int) {
	capacity := v.capacity()
	nbuf := make([]T, new_size)
	for i := 0; i < v.size; i += 1 {
		nbuf[i] = v.buf[(i+v.start)%capacity]
	}
	v.buf = nbuf
	v.start = 0
}
