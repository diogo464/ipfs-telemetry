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

var _ telemetry.Collector = (*kademliaQueryCollector)(nil)
var _ measurements.Kademlia = (*kademliaQueryCollector)(nil)

type kademliaQueryTiming struct {
	p peer.ID
	t datapoint.KademliaMessageType
	d time.Duration
}

type kademliaQueryCollector struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	queries []kademliaQueryTiming

	cquery chan kademliaQueryTiming
}

func KademliaQuery() telemetry.Collector {
	ctx, cancel := context.WithCancel(context.Background())
	c := &kademliaQueryCollector{
		ctx:     ctx,
		cancel:  cancel,
		queries: make([]kademliaQueryTiming, 0),
		cquery:  make(chan kademliaQueryTiming, 128),
	}
	measurements.KademliaRegister(c)
	go c.chanLoop()
	return c
}

// IncMessageIn implements measurements.Kademlia
func (*kademliaQueryCollector) IncMessageIn(pb.Message_MessageType) {
}

// IncMessageOut implements measurements.Kademlia
func (*kademliaQueryCollector) IncMessageOut(pb.Message_MessageType) {
}

// PushHandler implements measurements.Kademlia
func (*kademliaQueryCollector) PushHandler(p peer.ID, m pb.Message_MessageType, handler time.Duration, write time.Duration) {
}

// PushQuery implements measurements.Kademlia
func (c *kademliaQueryCollector) PushQuery(p peer.ID, t pb.Message_MessageType, d time.Duration) {
	c.cquery <- kademliaQueryTiming{
		p: p,
		t: ConvertKademliaMessageType(t),
		d: d,
	}
}

// Close implements telemetry.Collector
func (c *kademliaQueryCollector) Close() {
	c.cancel()
}

// Collect implements telemetry.Collector
func (c *kademliaQueryCollector) Collect(_ context.Context, stream *telemetry.Stream) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dp := &datapoint.KademliaQuery{}

	dp.Timestamp = datapoint.NewTimestamp()
	for _, q := range c.queries {
		dp.Peer = q.p
		dp.QueryType = q.t
		dp.Duration = q.d
		err := datapoint.KademliaQuerySerialize(dp, stream)
		if err != nil {
			return err
		}
	}
	c.queries = make([]kademliaQueryTiming, 0, len(c.queries))

	return nil
}

// Name implements telemetry.Collector
func (*kademliaQueryCollector) Name() string {
	return "Kademlia Queries"
}

func (c *kademliaQueryCollector) chanLoop() {
	for {
		select {
		case q := <-c.cquery:
			c.mu.Lock()
			c.queries = append(c.queries, q)
			c.mu.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
}
