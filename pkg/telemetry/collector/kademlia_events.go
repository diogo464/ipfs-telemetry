package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/measurements"
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ measurements.Kademlia = (*kademliaEventCollector)(nil)

type kademliaQueryTiming struct {
	p peer.ID
	t snapshot.KademliaMessageType
	d time.Duration
}

type kademliaHandlerTiming struct {
	p peer.ID
	t snapshot.KademliaMessageType
	h time.Duration
	w time.Duration
}

type kademliaEventCollector struct {
	cquery   chan kademliaQueryTiming
	chandler chan kademliaHandlerTiming
}

func StartKademliaEventCollector(ctx context.Context, sink snapshot.Sink) {
	c := &kademliaEventCollector{
		cquery:   make(chan kademliaQueryTiming),
		chandler: make(chan kademliaHandlerTiming),
	}
	measurements.KademliaRegister(c)
	go c.eventLoop(ctx, sink)
}

func (c *kademliaEventCollector) eventLoop(ctx context.Context, sink snapshot.Sink) {
LOOP:
	for {
		select {
		case timing := <-c.cquery:
			sink.Push(&snapshot.KademliaQuery{
				Timestamp: snapshot.NewTimestamp(),
				Peer:      timing.p,
				QueryType: timing.t,
				Duration:  timing.d,
			})
		case timing := <-c.chandler:
			sink.Push(&snapshot.KademliaHandler{
				Timestamp:       snapshot.NewTimestamp(),
				HandlerType:     timing.t,
				HandlerDuration: timing.h,
				WriteDuration:   timing.w,
			})
		case <-ctx.Done():
			// TODO: measurements unregister
			break LOOP
		}
	}
}

// IncMessageIn implements measurements.Kademlia
func (*kademliaEventCollector) IncMessageIn(snapshot.KademliaMessageType) {
}

// IncMessageOut implements measurements.Kademlia
func (*kademliaEventCollector) IncMessageOut(snapshot.KademliaMessageType) {
}

// PushHandler implements measurements.Kademlia
func (c *kademliaEventCollector) PushHandler(p peer.ID, m snapshot.KademliaMessageType, handler time.Duration, write time.Duration) {
	c.chandler <- kademliaHandlerTiming{
		p: p,
		t: m,
		h: handler,
		w: write,
	}
}

// PushQuery implements measurements.Kademlia
func (c *kademliaEventCollector) PushQuery(p peer.ID, t snapshot.KademliaMessageType, d time.Duration) {
	c.cquery <- kademliaQueryTiming{
		p: p,
		t: t,
		d: d,
	}
}
