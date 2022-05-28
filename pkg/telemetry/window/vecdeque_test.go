package window

import "testing"

func TestVecDeque(t *testing.T) {
	v := newVecDeque[int]()

	v.PushBack(1)
	v.PushBack(2)
	v.PushBack(3)
	v.PushBack(4)

	assertEq(t, v.Len(), 4)

	assertEq(t, v.PopFront(), 1)
	assertEq(t, v.Len(), 3)

	assertEq(t, v.PopFront(), 2)
	assertEq(t, v.Len(), 2)

	assertEq(t, v.Front(), 3)
	assertEq(t, v.Back(), 4)
	v.PushBack(5)
	assertEq(t, v.Back(), 5)
	assertEq(t, v.Len(), 3)

	assertEq(t, v.Front(), 3)
	assertEq(t, v.Back(), 5)
	assertEq(t, v.PopFront(), 3)
	assertEq(t, v.Len(), 2)

	assertEq(t, v.PopFront(), 4)
	assertEq(t, v.Len(), 1)

	assertEq(t, v.PopFront(), 5)
	assertEq(t, v.Len(), 0)
}

func assertEq(t *testing.T, expected int, received int) {
	if expected != received {
		t.Error("expected ", expected, " received ", received)
	}
}
