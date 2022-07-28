package ttlmap_test

import (
	"testing"
	"time"

	"git.d464.sh/uni/telemetry/ttlmap"
)

func TestTTLMap(t *testing.T) {
	m := ttlmap.New[string, int]()

	m.Insert("k1", 10, time.Millisecond*10)
	m.Insert("k2", 5, time.Millisecond*5)
	m.Insert("k3", 20, time.Millisecond*20)
	m.Insert("k4", 3, time.Millisecond*3)
	m.Insert("k5", 40, time.Millisecond*20)

	ensureExists := func(keys ...string) {
		for _, k := range keys {
			if !m.Contains(k) {
				t.FailNow()
			}
		}
	}
	ensureNotExists := func(keys ...string) {
		for _, k := range keys {
			if m.Contains(k) {
				t.FailNow()
			}
		}
	}

	ensureExists("k1", "k2", "k3", "k4", "k5")

	time.Sleep(time.Millisecond * 4)
	ensureExists("k1", "k2", "k3", "k5")
	ensureNotExists("k4")

	time.Sleep(time.Millisecond * 2)
	ensureExists("k1", "k3", "k5")
	ensureNotExists("k2", "k4")

	time.Sleep(time.Millisecond * 6)
	ensureExists("k3", "k5")
	ensureNotExists("k1", "k2", "k4")

	m.Insert("k5", 40, time.Millisecond*10)
	time.Sleep(time.Millisecond * 8)
	ensureExists("k5")
	ensureNotExists("k1", "k2", "k3", "k4")

	time.Sleep(time.Millisecond * 4)
	ensureNotExists("k1", "k2", "k3", "k4", "k5")
}
