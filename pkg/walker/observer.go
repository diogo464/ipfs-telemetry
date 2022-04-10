package walker

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Peer struct {
	ID        peer.ID
	Addresses []multiaddr.Multiaddr
	Agent     string
	Protocols []string
	Buckets   [][]peer.AddrInfo
	Requests  []Request
}

type Request struct {
	Start    time.Time
	Duration time.Duration
}

type Error struct {
	ID        peer.ID
	Addresses []multiaddr.Multiaddr
	Time      time.Time
	Err       error
}

type Observer interface {
	ObservePeer(*Peer)
	ObserveError(*Error)
}

var _ Observer = (*PeerObserverFn)(nil)

type PeerObserverFn struct {
	fn func(*Peer)
}

func NewPeerObserverFn(fn func(*Peer)) *PeerObserverFn {
	return &PeerObserverFn{fn}
}

// ObserveError implements Observer
func (*PeerObserverFn) ObserveError(*Error) {
}

// ObservePeer implements Observer
func (o *PeerObserverFn) ObservePeer(p *Peer) {
	(o.fn)(p)
}

var _ Observer = (*MultiObserver)(nil)

type MultiObserver struct {
	observers []Observer
}

func NewMultiObserver(observers ...Observer) *MultiObserver {
	arr := make([]Observer, 0, len(observers))
	arr = append(arr, observers...)
	return &MultiObserver{observers: arr}
}

// ObserveError implements Observer
func (o *MultiObserver) ObserveError(e *Error) {
	for _, obs := range o.observers {
		obs.ObserveError(e)
	}
}

// ObservePeer implements Observer
func (o *MultiObserver) ObservePeer(p *Peer) {
	for _, obs := range o.observers {
		obs.ObservePeer(p)
	}
}

var _ Observer = (*NullObserver)(nil)

type NullObserver struct {
}

// ObserveError implements Observer
func (*NullObserver) ObserveError(*Error) {
}

// ObservePeer implements Observer
func (*NullObserver) ObservePeer(*Peer) {
}
