package collector

import (
	"context"
	"fmt"
	"net"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"git.d464.sh/adc/telemetry/pkg/traceroute"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type TraceRouteOptions struct {
	Interval time.Duration
}

type traceRouteCollector struct {
	ctx  context.Context
	opts TraceRouteOptions
	sink snapshot.Sink
	h    host.Host
}

func RunTraceRouteCollector(ctx context.Context, h host.Host, sink snapshot.Sink, opts TraceRouteOptions) {
	c := &traceRouteCollector{ctx, opts, sink, h}
	c.Run()
}

func (c *traceRouteCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
	picker := newPeerPicker(c.h)
	defer picker.close()

LOOP:
	for {
		select {
		case <-ticker.C:
			if p, ok := picker.pick(); ok {
				if s, err := c.trace(p); err == nil {
					c.sink.Push(s)
				}
			}
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func (c *traceRouteCollector) trace(p peer.ID) (*snapshot.TraceRoute, error) {
	ip, ok := getFirstPublicIPFromMultiaddrs(c.h.Peerstore().Addrs(p))
	if !ok {
		return nil, fmt.Errorf("peer has no public IP's")
	}
	result, err := traceroute.Trace(ip)
	if err != nil {
		return nil, err
	}
	origin := peer.AddrInfo{ID: c.h.ID(), Addrs: c.h.Addrs()}
	destination := c.h.Peerstore().PeerInfo(p)

	return &snapshot.TraceRoute{
		Timestamp:   snapshot.NewTimestamp(),
		Origin:      origin,
		Destination: destination,
		Provider:    result.Provider,
		Output:      result.Output,
	}, nil
}

func getFirstPublicIPFromMultiaddrs(in []multiaddr.Multiaddr) (string, bool) {
	checkPrivate := func(in string) bool {
		if ip := net.ParseIP(in); ip != nil {
			return ip.IsPrivate()
		}
		return true
	}

	for _, addr := range in {
		for _, code := range []int{multiaddr.P_IP4, multiaddr.P_IP6} {
			if v, err := addr.ValueForProtocol(code); err == nil {
				if !checkPrivate(v) {
					return v, true
				}
			}
		}
	}
	return "", false
}
