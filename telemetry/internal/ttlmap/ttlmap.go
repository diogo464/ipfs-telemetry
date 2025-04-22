package ttlmap

import (
	"container/heap"
	"time"
)

type mapEntry[V any] struct {
	value      V
	expiration time.Time
}

type heapEntry[K comparable] struct {
	key        *K
	expiration time.Time
}

type mapHeap[K comparable] []heapEntry[K]

func (pq mapHeap[K]) Len() int { return len(pq) }

func (pq mapHeap[K]) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].expiration.Before(pq[j].expiration)
}

func (pq mapHeap[K]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *mapHeap[K]) Push(x any) {
	item := x.(heapEntry[K])
	*pq = append(*pq, item)
}

func (pq *mapHeap[K]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1].key = nil
	*pq = old[0 : n-1]
	return item
}

type Map[K comparable, V any] struct {
	entries map[K]*mapEntry[V]
	expheap mapHeap[K]
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		entries: make(map[K]*mapEntry[V]),
		expheap: make([]heapEntry[K], 0),
	}
}

func (m *Map[K, V]) Insert(key K, value V, ttl time.Duration) {
	m.clean()
	expiration := time.Now().Add(ttl)
	m.entries[key] = &mapEntry[V]{
		value:      value,
		expiration: expiration,
	}
	hentry := heapEntry[K]{
		key:        &key,
		expiration: expiration,
	}
	heap.Push(&m.expheap, hentry)
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	m.clean()
	if entry, ok := m.entries[key]; ok {
		return entry.value, true
	} else {
		var value V
		return value, false
	}
}

func (m *Map[K, V]) Remove(key K) (V, bool) {
	m.clean()
	var value V
	var present bool = false
	if entry, ok := m.entries[key]; ok {
		value = entry.value
		present = true
	}
	delete(m.entries, key)
	return value, present
}

func (m *Map[K, V]) Contains(key K) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *Map[K, V]) clean() {
	now := time.Now()
	for len(m.expheap) > 0 {
		hentry := m.expheap[0]
		mentry, ok := m.entries[*hentry.key]
		if ok && mentry.expiration != hentry.expiration {
			m.expheap[0].expiration = mentry.expiration
			heap.Fix(&m.expheap, 0)
			continue
		}
		if hentry.expiration.Before(now) {
			delete(m.entries, *hentry.key)
			heap.Pop(&m.expheap)
		} else {
			break
		}
	}
}
