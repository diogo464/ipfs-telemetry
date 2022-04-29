package telemetry

import (
	"net"
	"sync"
	"time"
)

type serviceThrottler struct {
	blocked_mu sync.Mutex
	blocked    map[string]time.Time
}

func newServiceThrottler() *serviceThrottler {
	return &serviceThrottler{
		blocked_mu: sync.Mutex{},
		blocked:    make(map[string]time.Time),
	}
}

func (t *serviceThrottler) isAllowed(ip net.IP) bool {
	t.blocked_mu.Lock()
	defer t.blocked_mu.Unlock()
	t.clean()
	_, ok := t.blocked[ip.String()]
	return !ok
}

func (t *serviceThrottler) disallow(ip net.IP, dur time.Duration) {
	t.blocked_mu.Lock()
	defer t.blocked_mu.Unlock()
	t.blocked[ip.String()] = time.Now().Add(dur)
}

func (t *serviceThrottler) clean() {
	now := time.Now()
	for k, v := range t.blocked {
		if !v.After(now) {
			delete(t.blocked, k)
		}
	}
}
