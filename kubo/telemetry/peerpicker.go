package telemetry

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type peerPicker struct {
	sync.Mutex
	h        host.Host
	picked   map[peer.ID]struct{}
	unpicked map[peer.ID]struct{}
}

func newPeerPicker(h host.Host) *peerPicker {
	picker := &peerPicker{
		h:        h,
		picked:   make(map[peer.ID]struct{}),
		unpicked: make(map[peer.ID]struct{}),
	}
	h.Network().Notify(picker)
	return picker
}

func (c *peerPicker) pick() (peer.ID, bool) {
	c.Lock()
	defer c.Unlock()
	if len(c.picked) == 0 && len(c.unpicked) != 0 {
		c.picked, c.unpicked = c.unpicked, c.picked
	}
	for v := range c.unpicked {
		delete(c.unpicked, v)
		c.picked[v] = struct{}{}
		return v, true
	}
	return peer.ID(""), false
}

func (c *peerPicker) close() {
	c.h.Network().StopNotify(c)
}

// network.Notifiee impl
// called when network starts listening on an addr
func (c *peerPicker) Listen(network.Network, multiaddr.Multiaddr) {}

// called when network stops listening on an addr
func (c *peerPicker) ListenClose(network.Network, multiaddr.Multiaddr) {}

// called when a connection opened
func (c *peerPicker) Connected(n network.Network, conn network.Conn) {
	c.Lock()
	defer c.Unlock()
	c.unpicked[conn.RemotePeer()] = struct{}{}
}

// called when a connection closed
func (c *peerPicker) Disconnected(n network.Network, conn network.Conn) {
	c.Lock()
	defer c.Unlock()
	delete(c.picked, conn.RemotePeer())
	delete(c.unpicked, conn.RemotePeer())
}

// called when a stream opened
func (c *peerPicker) OpenedStream(network.Network, network.Stream) {}

// called when a stream closed
func (c *peerPicker) ClosedStream(network.Network, network.Stream) {}
