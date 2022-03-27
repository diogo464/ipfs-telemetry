package snapshot

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

type Ping struct {
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func NewPing(src peer.AddrInfo, dest peer.AddrInfo, durs []time.Duration) *Snapshot {
	return NewSnapshot("ping", &Ping{
		Source:      src,
		Destination: dest,
		Durations:   durs,
	})
}

type PingOptions struct {
	PingCount int
	Interval  time.Duration
	Timeout   time.Duration
}

type PingCollector struct {
	opts PingOptions
	h    host.Host
	sink Sink
}

func NewPingCollector(h host.Host, sink Sink, opts PingOptions) *PingCollector {
	return &PingCollector{opts: opts, h: h, sink: sink}
}

func (c *PingCollector) Run() {
	for {
		if peerid, ok := c.pickRandomPeer(); !ok {
			time.Sleep(time.Second)
			continue
		} else {
			snapshot, err := c.ping(peerid)
			if err == nil {
				fmt.Println(snapshot)
				c.sink.Push(snapshot)
			} else {
				fmt.Println(err)
			}
			time.Sleep(c.opts.Interval)
		}
	}
}

func (c *PingCollector) pickRandomPeer() (peer.ID, bool) {
	peers := c.h.Peerstore().PeersWithAddrs()
	lpeers := len(peers)
	if lpeers == 0 {
		return peer.ID(""), false
	}
	index := rand.Intn(lpeers)
	peerid := peers[index]
	return peerid, true
}

func (c *PingCollector) ping(p peer.ID) (*Snapshot, error) {
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

	return NewPing(source, destination, durations), nil
}
