package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/measurements"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

type KademliaOptions struct {
	Interval time.Duration
}

type kademliaCollectorTiming struct {
	p peer.ID
	d time.Duration
}

type kademliaCollector struct {
	ctx  context.Context
	opts KademliaOptions
	sink snapshot.Sink

	ctiming chan kademliaCollectorTiming
}

func RunKademliaCollector(ctx context.Context, sink snapshot.Sink, opts KademliaOptions) {
	c := &kademliaCollector{
		ctx:  ctx,
		opts: opts,
		sink: sink,

		ctiming: make(chan kademliaCollectorTiming, 128),
	}
	c.Run()
}

func (c *kademliaCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
	measurements.KademliaRegister(c)

LOOP:
	for {
		select {
		case <-ticker.C:
		case timing := <-c.ctiming:
			c.sink.Push(&snapshot.KademliaQuery{
				Timestamp: snapshot.NewTimestamp(),
				Peer:      timing.p,
				Duration:  timing.d,
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// measurements.Kademlia impl
func (c *kademliaCollector) PushQueryTiming(p peer.ID, d time.Duration) {
	c.ctiming <- kademliaCollectorTiming{
		p: p,
		d: d,
	}
}
