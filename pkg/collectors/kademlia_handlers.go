package collectors

import (
	"context"
	"sync"
	"time"

	"github.com/diogo464/telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

var _ telemetry.Collector = (*kademliaHandlerCollector)(nil)
var _ measurements.Kademlia = (*kademliaHandlerCollector)(nil)

type kademliaHandlerTiming struct {
	p peer.ID
	t datapoint.KademliaMessageType
	h time.Duration
	w time.Duration
}

type kademliaHandlerCollector struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	handlers []kademliaHandlerTiming

	chandler chan kademliaHandlerTiming
}

func KademliaHandler() telemetry.Collector {
	ctx, cancel := context.WithCancel(context.Background())
	c := &kademliaHandlerCollector{
		ctx:      ctx,
		cancel:   cancel,
		handlers: make([]kademliaHandlerTiming, 0),
		chandler: make(chan kademliaHandlerTiming, 128),
	}
	measurements.KademliaRegister(c)
	go c.chanLoop()
	return c
}

// IncMessageIn implements measurements.Kademlia
func (*kademliaHandlerCollector) IncMessageIn(pb.Message_MessageType) {
}

// IncMessageOut implements measurements.Kademlia
func (*kademliaHandlerCollector) IncMessageOut(pb.Message_MessageType) {
}

// PushHandler implements measurements.Kademlia
func (c *kademliaHandlerCollector) PushHandler(p peer.ID, m pb.Message_MessageType, handler time.Duration, write time.Duration) {
	c.chandler <- kademliaHandlerTiming{
		p: p,
		t: ConvertKademliaMessageType(m),
		h: handler,
		w: write,
	}
}

// PushQuery implements measurements.Kademlia
func (*kademliaHandlerCollector) PushQuery(p peer.ID, t pb.Message_MessageType, d time.Duration) {
}

// Close implements telemetry.Collector
func (c *kademliaHandlerCollector) Close() {
	c.cancel()
}

// Collect implements telemetry.Collector
func (c *kademliaHandlerCollector) Collect(_ context.Context, stream *telemetry.Stream) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dp := &datapoint.KademliaHandler{}

	dp.Timestamp = datapoint.NewTimestamp()
	for _, h := range c.handlers {
		dp.HandlerType = h.t
		dp.HandlerDuration = h.h
		dp.WriteDuration = h.w
		err := datapoint.KademliaHandlerSerialize(dp, stream)
		if err != nil {
			return err
		}
	}
	c.handlers = make([]kademliaHandlerTiming, 0, len(c.handlers))

	return nil
}

// Name implements telemetry.Collector
func (*kademliaHandlerCollector) Name() string {
	return "Kademlia Handlers"
}

func (c *kademliaHandlerCollector) chanLoop() {
	for {
		select {
		case q := <-c.chandler:
			c.mu.Lock()
			c.handlers = append(c.handlers, q)
			c.mu.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
}
