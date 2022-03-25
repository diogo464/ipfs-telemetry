package window

type windowItem struct {
	key   uint64
	value interface{}
}

type window struct {
	stamp uint64
	size  int
	begin int
	items []windowItem
}

func newWindow() *window {
	return &window{
		stamp: 0,
		size:  0,
		begin: 0,
		items: []windowItem{},
	}
}

func (w *window) push(item windowItem) {
	if w.size == w.capacity() {
		w.reallocate(w.capacity())
	}
	index := w.begin + w.size
	w.size += 1
	w.items[index] = item
}

func (w *window) pop() {
	if w.isEmpty() {
		return
	}
	w.items[w.begin].value = nil
	w.begin = (w.begin + 1) % w.capacity()
}

func (w *window) findFirstGte(key uint64) int {
	left := 0
	right := w.size
	for left != right {
		mid := (left + right) / 2
		midkey := w.get(mid).key
		if midkey >= key {
			right = mid
		} else if midkey < key {
			left = mid
		}
	}
	return left
}

func (w *window) get(index int) windowItem {
	if index > w.size || index < 0 {
		panic("invalid index")
	}
	return w.items[(w.begin+index)%w.capacity()]
}

func (w *window) first() windowItem {
	return w.items[w.begin]
}

func (w *window) isEmpty() bool {
	return w.size == 0
}

func (w *window) capacity() int {
	if len(w.items) != cap(w.items) {
		panic("?")
	}
	return len(w.items)
}

func (w *window) reallocate(size int) {
	if size < w.size {
		panic("invalid size")
	}

	items := make([]windowItem, size)
	for i := 0; i < w.size; i++ {
		items[i] = w.items[i+w.begin]
	}

	w.begin = 0
	w.items = items
}

func (w *window) popUpTo(key uint64) {
	for !w.isEmpty() && w.first().key < key {
		w.pop()
	}
}

func saturatingSub(a, b uint64) uint64 {
	if a < b {
		return 0
	}
	return a - b
}
