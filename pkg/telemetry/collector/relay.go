package collector

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/atomic"
)

var _ Collector = (*RelayCollector)(nil)
var _ measurements.Relay = (*RelayCollector)(nil)

type connectionClose struct {
	initiator peer.ID
	target    peer.ID
	duration  time.Duration
	relayed   uint64
}

type connectionOpen struct {
	initiator peer.ID
	target    peer.ID
}

type reservation struct {
	p peer.ID
}

type RelayCollector struct {
	ctx    context.Context
	cancel context.CancelFunc

	eventSink datapoint.Sink

	cconnectionClose chan connectionClose
	cconectionOpen   chan connectionOpen
	creservation     chan reservation

	reservations      *atomic.Uint32
	connections       *atomic.Uint32
	bytesRelayed      *atomic.Uint64
	activeConnections *atomic.Uint32
}

func NewRelayCollector() *RelayCollector {
	ctx, cancel := context.WithCancel(context.Background())
	c := &RelayCollector{
		ctx:    ctx,
		cancel: cancel,

		eventSink: nil,

		cconnectionClose: make(chan connectionClose, 128),
		cconectionOpen:   make(chan connectionOpen, 128),
		creservation:     make(chan reservation, 128),

		reservations:      atomic.NewUint32(0),
		connections:       atomic.NewUint32(0),
		bytesRelayed:      atomic.NewUint64(0),
		activeConnections: atomic.NewUint32(0),
	}
	measurements.RelayRegister(c)
	go c.eventLoop()
	return c
}

func (c *RelayCollector) SetEventSink(sink datapoint.Sink) {
	c.eventSink = sink
}

// Close implements Collector
func (c *RelayCollector) Close() {
	// TODO: unregister relay
	c.cancel()
}

// Collect implements Collector
func (c *RelayCollector) Collect(_ context.Context, sink datapoint.Sink) {
	sink.Push(&datapoint.RelayStats{
		Timestamp:         datapoint.NewTimestamp(),
		Reservations:      c.reservations.Load(),
		Connections:       c.connections.Load(),
		BytesRelayed:      c.bytesRelayed.Load(),
		ActiveConnections: c.activeConnections.Load(),
	})
}

// ConnectionClose implements measurements.Relay
func (c *RelayCollector) ConnectionClose(initiator peer.ID, target peer.ID, duration time.Duration, relayed uint64) {
	c.cconnectionClose <- connectionClose{
		initiator: initiator,
		target:    target,
		duration:  duration,
		relayed:   relayed,
	}
}

// ConnectionOpen implements measurements.Relay
func (c *RelayCollector) ConnectionOpen(initiator peer.ID, target peer.ID) {
	c.cconectionOpen <- connectionOpen{
		initiator: initiator,
		target:    target,
	}
}

// Reservation implements measurements.Relay
func (c *RelayCollector) Reservation(p peer.ID) {
	c.creservation <- reservation{p: p}
}

func (c *RelayCollector) eventLoop() {
LOOP:
	for {
		select {
		case ev := <-c.cconnectionClose:
			c.activeConnections.Dec()
			c.pushEvent(&datapoint.RelayComplete{
				Timestamp:    datapoint.NewTimestamp(),
				Duration:     ev.duration,
				Initiator:    ev.initiator,
				Target:       ev.target,
				BytesRelayed: ev.relayed,
			})
		case ev := <-c.cconectionOpen:
			c.activeConnections.Inc()
			c.pushEvent(&datapoint.RelayConnection{
				Timestamp: datapoint.NewTimestamp(),
				Initiator: ev.initiator,
				Target:    ev.target,
			})
		case ev := <-c.creservation:
			c.reservations.Inc()
			c.pushEvent(&datapoint.RelayReservation{
				Timestamp: datapoint.NewTimestamp(),
				Peer:      ev.p,
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func (c *RelayCollector) pushEvent(ev datapoint.Datapoint) {
	if c.eventSink != nil {
		c.eventSink.Push(ev)
	}
}
