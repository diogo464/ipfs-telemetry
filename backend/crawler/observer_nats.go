package crawler

import (
	"encoding/json"
	"time"

	"github.com/diogo464/ipfs-telemetry/backend/monitor"
	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/crawler"
	"github.com/diogo464/telemetry/walker"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var _ (crawler.Observer) = (*natsObserver)(nil)

type natsObserver struct {
	l  *zap.Logger
	nc *nats.Conn
}

func newNatsObserver(l *zap.Logger, natsUrl string) (*natsObserver, error) {
	l.Info("connecting to nats at " + natsUrl)
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		l.Error("failed to connect to nats at "+natsUrl, zap.Error(err))
		return nil, err
	}
	return &natsObserver{
		l:  l,
		nc: nc,
	}, nil
}

func (o *natsObserver) CrawlBegin() {
	o.publishMessage(NatsMessage{
		Kind:      KindCrawlBegin,
		Timestamp: time.Now(),
	})
}

func (o *natsObserver) CrawlEnd() {
	o.publishMessage(NatsMessage{
		Kind:      KindCrawlEnd,
		Timestamp: time.Now(),
	})
}

func (o *natsObserver) ObservePeer(c *walker.Peer) {
	o.publishMessage(NatsMessage{
		Kind:      KindPeer,
		Timestamp: time.Now(),
		Peer:      c,
	})

	if c.ContainsProtocol(telemetry.ID_TELEMETRY) {
		if m, err := walkerPeerToDiscoveryMarshaled(c); err == nil {
			if err := o.nc.Publish(monitor.Subject_Discover, m); err != nil {
				o.l.Error("failed to publish discovery message", zap.String("subject", monitor.Subject_Discover), zap.Error(err))
			}
		} else {
			o.l.Error("failed to marshal discovery", zap.Error(err))
		}
	}
}

func (*natsObserver) ObserveError(*walker.Error) {
}

func (o *natsObserver) publishMessage(msg NatsMessage) {
	m, err := json.Marshal(&msg)
	if err != nil {
		o.l.Error("failed to marshal message", zap.Error(err))
		return
	}

	if err := o.nc.Publish(SubjectCrawler, m); err != nil {
		o.l.Error("failed to publish message", zap.Error(err))
		return
	}
}

func walkerPeerToDiscovery(c *walker.Peer) monitor.DiscoveryMessage {
	return monitor.DiscoveryMessage{
		ID:        c.ID,
		Addresses: c.Addresses,
	}
}

func walkerPeerToDiscoveryMarshaled(c *walker.Peer) ([]byte, error) {
	d := walkerPeerToDiscovery(c)
	m, err := json.Marshal(d)
	return m, err
}
