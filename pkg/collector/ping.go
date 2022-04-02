package collector

import (
	"context"
	"math/rand"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type PingOptions struct {
	PingCount int
	Interval  time.Duration
	Timeout   time.Duration
}

type pingResult struct {
	ps  *snapshot.Ping
	err error
}

type pingCollector struct {
	ctx  context.Context
	opts PingOptions
	h    host.Host
	sink snapshot.Sink

	cresult       chan pingResult
	cconnected    chan peer.ID
	cdisconnected chan peer.ID
}

func RunPingCollector(ctx context.Context, h host.Host, sink snapshot.Sink, opts PingOptions) {
	c := &pingCollector{
		ctx:           ctx,
		opts:          opts,
		h:             h,
		sink:          sink,
		cresult:       make(chan pingResult),
		cconnected:    make(chan peer.ID),
		cdisconnected: make(chan peer.ID),
	}
	c.Run()
}

func (c *pingCollector) Run() {
	c.h.Network().Notify(c)
	ticker := time.NewTicker(c.opts.Interval)

	inprogress := false
	pending := make(map[peer.ID]struct{})
	completed := make(map[peer.ID]struct{})

LOOP:
	for {
		select {
		case p := <-c.cconnected:
			pending[p] = struct{}{}
		case p := <-c.cdisconnected:
			delete(pending, p)
			delete(completed, p)
		case r := <-c.cresult:
			inprogress = false
			if r.err == nil {
				c.sink.PushPing(r.ps)
			}
		case <-ticker.C:
			if !inprogress {
				for p := range pending {
					inprogress = true
					go func() {
						ps, err := c.ping(p)
						c.cresult <- pingResult{ps, err}
					}()
					break
				}
			}
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func (c *pingCollector) pickRandomPeer() (peer.ID, bool) {
	peers := c.h.Peerstore().PeersWithAddrs()
	lpeers := len(peers)
	if lpeers == 0 {
		return peer.ID(""), false
	}
	index := rand.Intn(lpeers)
	peerid := peers[index]
	return peerid, true
}

func (c *pingCollector) ping(p peer.ID) (*snapshot.Ping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opts.Timeout)
	defer cancel()

	if c.h.Network().Connectedness(p) != network.Connected {
		if err := c.h.Connect(ctx, c.h.Peerstore().PeerInfo(p)); err != nil {
			return nil, err
		}
	}

	durations := make([]time.Duration, c.opts.PingCount)
	counter := 0
	cresult := ping.Ping(network.WithNoDial(ctx, "ping"), c.h, p)
	for result := range cresult {
		if result.Error != nil {
			return nil, result.Error
		}
		durations[counter] = result.RTT
		counter += 1
		if counter == c.opts.PingCount {
			break
		}
	}

	source := peer.AddrInfo{
		ID:    c.h.ID(),
		Addrs: c.h.Addrs(),
	}
	destination := c.h.Peerstore().PeerInfo(p)

	return &snapshot.Ping{
		Timestamp:   time.Now().UTC(),
		Source:      source,
		Destination: destination,
		Durations:   durations,
	}, nil
}

// network.Notifiee impl
// called when network starts listening on an addr
func (c *pingCollector) Listen(network.Network, multiaddr.Multiaddr) {}

// called when network stops listening on an addr
func (c *pingCollector) ListenClose(network.Network, multiaddr.Multiaddr) {}

// called when a connection opened
func (c *pingCollector) Connected(n network.Network, conn network.Conn) {
	c.cconnected <- conn.RemotePeer()
}

// called when a connection closed
func (c *pingCollector) Disconnected(n network.Network, conn network.Conn) {
	c.cdisconnected <- conn.RemotePeer()
}

// called when a stream opened
func (c *pingCollector) OpenedStream(network.Network, network.Stream) {}

// called when a stream closed
func (c *pingCollector) ClosedStream(network.Network, network.Stream) {}
