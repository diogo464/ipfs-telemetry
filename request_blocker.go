package telemetry

import (
	"net"
	"sync"
	"time"

	"github.com/diogo464/telemetry/ttlmap"
)

type requestBlocker struct {
	mu     sync.Mutex
	blocks *ttlmap.Map[string, struct{}]
}

func newRequestBlocker() *requestBlocker {
	return &requestBlocker{
		blocks: ttlmap.New[string, struct{}](),
	}
}

func (r *requestBlocker) isBlocked(ip net.IP) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.blocks.Contains(ip.String())
}

func (r *requestBlocker) block(ip net.IP, dur time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blocks.Insert(ip.String(), struct{}{}, dur)
}
