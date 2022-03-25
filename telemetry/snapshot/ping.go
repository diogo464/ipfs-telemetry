package snapshot

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

type PingSnapshot struct {
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func NewPingSnapshot(src peer.AddrInfo, dest peer.AddrInfo, durs []time.Duration) *Snapshot {
	return NewSnapshot("ping", &PingSnapshot{
		Source:      src,
		Destination: dest,
		Durations:   durs,
	})
}
