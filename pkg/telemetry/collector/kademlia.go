package collector

import (
	"context"
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/measurements"
	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Collector = (*kademliaCollector)(nil)
var _ measurements.Kademlia = (*kademliaCollector)(nil)

type kademliaCollector struct {
	ctx    context.Context
	cancel context.CancelFunc

	messages_mu  sync.Mutex
	messages_in  map[datapoint.KademliaMessageType]uint64
	messages_out map[datapoint.KademliaMessageType]uint64

	cquery   chan kademliaQueryTiming
	chandler chan kademliaHandlerTiming
	cmsgin   chan datapoint.KademliaMessageType
	cmsgout  chan datapoint.KademliaMessageType
}

func NewKademliaCollector() Collector {
	ctx, cancel := context.WithCancel(context.Background())
	c := &kademliaCollector{
		ctx:    ctx,
		cancel: cancel,

		messages_in:  make(map[datapoint.KademliaMessageType]uint64),
		messages_out: make(map[datapoint.KademliaMessageType]uint64),
	}
	go c.eventLoop()
	return c
}

// Close implements Collector
func (c *kademliaCollector) Close() {
	// TODO: measurements unregister
	c.cancel()
}

// Collect implements Collector
func (c *kademliaCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	c.messages_mu.Lock()
	defer c.messages_mu.Unlock()
	// TODO: clone map
	sink.Push(&datapoint.Kademlia{
		Timestamp:   datapoint.NewTimestamp(),
		MessagesIn:  c.messages_in,
		MessagesOut: c.messages_out,
	})
}

func (c *kademliaCollector) eventLoop() {
LOOP:
	for {
		select {
		case t := <-c.cmsgin:
			c.messages_mu.Lock()
			defer c.messages_mu.Unlock()
			c.messages_in[t] += 1
		case t := <-c.cmsgout:
			c.messages_mu.Lock()
			defer c.messages_mu.Unlock()
			c.messages_out[t] += 1
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// IncMessageIn implements measurements.Kademlia
func (c *kademliaCollector) IncMessageIn(t datapoint.KademliaMessageType) {
	c.cmsgin <- t
}

// IncMessageOut implements measurements.Kademlia
func (c *kademliaCollector) IncMessageOut(t datapoint.KademliaMessageType) {
	c.cmsgout <- t
}

// PushHandler implements measurements.Kademlia
func (c *kademliaCollector) PushHandler(p peer.ID, m datapoint.KademliaMessageType, handler time.Duration, write time.Duration) {
	c.chandler <- kademliaHandlerTiming{
		p: p,
		t: m,
		h: handler,
		w: write,
	}
}

// PushQuery implements measurements.Kademlia
func (c *kademliaCollector) PushQuery(p peer.ID, t datapoint.KademliaMessageType, d time.Duration) {
	c.cquery <- kademliaQueryTiming{
		p: p,
		t: t,
		d: d,
	}
}
