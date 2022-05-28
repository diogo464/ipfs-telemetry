package collector

import (
	"context"
	"sync"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
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

	cmsgin  chan datapoint.KademliaMessageType
	cmsgout chan datapoint.KademliaMessageType
}

func NewKademliaCollector() Collector {
	ctx, cancel := context.WithCancel(context.Background())
	c := &kademliaCollector{
		ctx:    ctx,
		cancel: cancel,

		messages_in:  make(map[datapoint.KademliaMessageType]uint64),
		messages_out: make(map[datapoint.KademliaMessageType]uint64),

		cmsgin:  make(chan datapoint.KademliaMessageType, 64),
		cmsgout: make(chan datapoint.KademliaMessageType, 64),
	}
	for _, ty := range datapoint.KademliaMessageTypes {
		c.messages_in[ty] = 0
		c.messages_out[ty] = 0
	}
	measurements.KademliaRegister(c)
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
	sink.Push(&datapoint.Kademlia{
		Timestamp:   datapoint.NewTimestamp(),
		MessagesIn:  cloneMap(c.messages_in),
		MessagesOut: cloneMap(c.messages_out),
	})
}

func (c *kademliaCollector) eventLoop() {
LOOP:
	for {
		select {
		case t := <-c.cmsgin:
			c.messages_mu.Lock()
			c.messages_in[t] += 1
			c.messages_mu.Unlock()
		case t := <-c.cmsgout:
			c.messages_mu.Lock()
			c.messages_out[t] += 1
			c.messages_mu.Unlock()
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
}

// PushQuery implements measurements.Kademlia
func (c *kademliaCollector) PushQuery(p peer.ID, t datapoint.KademliaMessageType, d time.Duration) {
}

func cloneMap(in map[datapoint.KademliaMessageType]uint64) map[datapoint.KademliaMessageType]uint64 {
	out := make(map[datapoint.KademliaMessageType]uint64)
	for k, v := range in {
		out[k] = v
	}
	return out
}
