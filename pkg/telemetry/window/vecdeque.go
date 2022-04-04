package window

type vecdeque[T any] struct {
	buf []T
}

func newVecDeque[T any]() *vecdeque[T] {
	return &vecdeque[T]{
		buf: []T{},
	}
}

func (v *vecdeque[T]) PushBack(x T) {
	v.buf = append(v.buf, x)
}

func (v *vecdeque[T]) PopFront() {
	v.buf = v.buf[1:]
}

func (v *vecdeque[T]) Get(i int) T {
	return v.buf[i]
}

func (v *vecdeque[T]) Back() T {
	return v.buf[len(v.buf)-1]
}

func (v *vecdeque[T]) Front() T {
	return v.buf[0]
}

func (v *vecdeque[T]) Len() int {
	return len(v.buf)
}

func (v *vecdeque[T]) IsEmpty() bool {
	return v.Len() == 0
}
