package collector

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ measurements.Kademlia = (*kademliaEventCollector)(nil)

type kademliaQueryTiming struct {
	p peer.ID
	t datapoint.KademliaMessageType
	d time.Duration
}

type kademliaHandlerTiming struct {
	p peer.ID
	t datapoint.KademliaMessageType
	h time.Duration
	w time.Duration
}

type kademliaEventCollector struct {
	cquery   chan kademliaQueryTiming
	chandler chan kademliaHandlerTiming
}

func StartKademliaEventCollector(ctx context.Context, sink datapoint.Sink) {
	c := &kademliaEventCollector{
		cquery:   make(chan kademliaQueryTiming),
		chandler: make(chan kademliaHandlerTiming),
	}
	measurements.KademliaRegister(c)
	go c.eventLoop(ctx, sink)
}

func (c *kademliaEventCollector) eventLoop(ctx context.Context, sink datapoint.Sink) {
LOOP:
	for {
		select {
		case timing := <-c.cquery:
			sink.Push(&datapoint.KademliaQuery{
				Timestamp: datapoint.NewTimestamp(),
				Peer:      timing.p,
				QueryType: timing.t,
				Duration:  timing.d,
			})
		case timing := <-c.chandler:
			sink.Push(&datapoint.KademliaHandler{
				Timestamp:       datapoint.NewTimestamp(),
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
func (*kademliaEventCollector) IncMessageIn(datapoint.KademliaMessageType) {
}

// IncMessageOut implements measurements.Kademlia
func (*kademliaEventCollector) IncMessageOut(datapoint.KademliaMessageType) {
}

// PushHandler implements measurements.Kademlia
func (c *kademliaEventCollector) PushHandler(p peer.ID, m datapoint.KademliaMessageType, handler time.Duration, write time.Duration) {
	c.chandler <- kademliaHandlerTiming{
		p: p,
		t: m,
		h: handler,
		w: write,
	}
}

// PushQuery implements measurements.Kademlia
func (c *kademliaEventCollector) PushQuery(p peer.ID, t datapoint.KademliaMessageType, d time.Duration) {
	c.cquery <- kademliaQueryTiming{
		p: p,
		t: t,
		d: d,
	}
}
