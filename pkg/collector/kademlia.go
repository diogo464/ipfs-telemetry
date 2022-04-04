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
	t snapshot.KademliaMessageType
	d time.Duration
}

type kademliaCollector struct {
	ctx  context.Context
	opts KademliaOptions
	sink snapshot.Sink

	ctiming chan kademliaCollectorTiming
	cmsgin  chan snapshot.KademliaMessageType
	cmsgout chan snapshot.KademliaMessageType
}

func RunKademliaCollector(ctx context.Context, sink snapshot.Sink, opts KademliaOptions) {
	c := &kademliaCollector{
		ctx:  ctx,
		opts: opts,
		sink: sink,

		ctiming: make(chan kademliaCollectorTiming, 128),
		cmsgin:  make(chan snapshot.KademliaMessageType, 128),
		cmsgout: make(chan snapshot.KademliaMessageType, 128),
	}
	c.Run()
}

func (c *kademliaCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
	measurements.KademliaRegister(c)

	messages_in := make(map[snapshot.KademliaMessageType]uint64)
	messages_out := make(map[snapshot.KademliaMessageType]uint64)

LOOP:
	for {
		select {
		case <-ticker.C:
			c.sink.Push(&snapshot.Kademlia{
				Timestamp:   snapshot.NewTimestamp(),
				MessagesIn:  messages_in,
				MessagesOut: messages_out,
			})
		case timing := <-c.ctiming:
			c.sink.Push(&snapshot.KademliaQuery{
				Timestamp: snapshot.NewTimestamp(),
				Peer:      timing.p,
				Duration:  timing.d,
			})
		case t := <-c.cmsgin:
			messages_in[t] += 1
		case t := <-c.cmsgout:
			messages_out[t] += 1
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// measurements.Kademlia impl
func (c *kademliaCollector) IncMessageIn(t snapshot.KademliaMessageType) {
	c.cmsgin <- t
}
func (c *kademliaCollector) IncMessageOut(t snapshot.KademliaMessageType) {
	c.cmsgout <- t
}
func (c *kademliaCollector) PushQuery(p peer.ID, t snapshot.KademliaMessageType, d time.Duration) {
	c.ctiming <- kademliaCollectorTiming{
		p: p,
		t: t,
		d: d,
	}
}
