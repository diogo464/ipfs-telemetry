package vecdeque

import "github.com/diogo464/ipfs_telemetry/pkg/utils"

type VecDeque[T any] struct {
	buf   []T
	size  int
	start int
}

func New[T any]() *VecDeque[T] {
	return &VecDeque[T]{
		buf:   make([]T, 0),
		size:  0,
		start: 0,
	}
}

func (v *VecDeque[T]) PushBack(x T) {
	if v.size == v.capacity() {
		v.reallocate(utils.Max(1, v.size*2))
	}
	v.buf[(v.start+v.size)%v.capacity()] = x
	v.size += 1
}

func (v *VecDeque[T]) PopFront() T {
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

func (v *VecDeque[T]) Get(i int) T {
	return v.buf[(v.start+i)%v.capacity()]
}

func (v *VecDeque[T]) Back() T {
	if v.IsEmpty() {
		panic("Back on empty")
	}
	return v.buf[(v.start+v.size-1)%v.capacity()]
}

func (v *VecDeque[T]) BackRef() *T {
	if v.IsEmpty() {
		panic("BackRef on empty")
	}
	return &v.buf[(v.start+v.size-1)%v.capacity()]
}

func (v *VecDeque[T]) Front() T {
	if v.IsEmpty() {
		panic("Front on empty")
	}
	return v.buf[v.start]
}

func (v *VecDeque[T]) Len() int {
	return v.size
}

func (v *VecDeque[T]) IsEmpty() bool {
	return v.Len() == 0
}

func (v *VecDeque[T]) capacity() int {
	return cap(v.buf)
}

func (v *VecDeque[T]) reallocate(new_size int) {
	capacity := v.capacity()
	nbuf := make([]T, new_size)
	for i := 0; i < v.size; i += 1 {
		nbuf[i] = v.buf[(i+v.start)%capacity]
	}
	v.buf = nbuf
	v.start = 0
}
