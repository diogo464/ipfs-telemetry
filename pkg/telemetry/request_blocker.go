package telemetry

import (
	"net"
	"sync"
	"time"

	"github.com/diogo464/telemetry/pkg/ttlmap"
)

type requestBlocker struct {
	mu       sync.Mutex
	services map[string]*ttlmap.Map[string, struct{}]
}

func newRequestBlocker() *requestBlocker {
	return &requestBlocker{
		services: make(map[string]*ttlmap.Map[string, struct{}]),
	}
}

func (r *requestBlocker) isBlocked(tag string, ip net.IP) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.services[tag]; ok {
		ips := ip.String()
		return m.Contains(ips)
	} else {
		return false
	}
}

func (r *requestBlocker) block(tag string, ip net.IP, dur time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ips := ip.String()
	if r.services[tag] == nil {
		r.services[tag] = ttlmap.New[string, struct{}]()
	}
	r.services[tag].Insert(ips, struct{}{}, dur)
}
