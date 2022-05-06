package walker

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type BucketEntry struct {
	ID    peer.ID               `json:"id"`
	Addrs []multiaddr.Multiaddr `json:"addresses"`
}

type Peer struct {
	ID              peer.ID               `json:"id"`
	Addresses       []multiaddr.Multiaddr `json:"addresses"`
	Agent           string                `json:"agent"`
	Protocols       []string              `json:"protocols"`
	Buckets         []BucketEntry         `json:"buckets"`
	Requests        []Request             `json:"requests"`
	ConnectStart    time.Time             `json:"connect_start"`
	ConnectDuration time.Duration         `json:"connect_duration"`
}

type Request struct {
	Start    time.Time     `json:"start"`
	Duration time.Duration `json:"duration"`
}

type Error struct {
	ID        peer.ID               `json:"id"`
	Addresses []multiaddr.Multiaddr `json:"addresses"`
	Time      time.Time             `json:"time"`
	Err       error                 `json:"error"`
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

var _ Observer = (*CollectorObserver)(nil)

type CollectorObserver struct {
	mu     sync.Mutex
	Peers  []*Peer  `json:"peers"`
	Errors []*Error `json:"errors"`
}

func NewCollectorObserver() *CollectorObserver {
	return &CollectorObserver{
		Peers:  []*Peer{},
		Errors: []*Error{},
	}
}

// ObserveError implements Observer
func (c *CollectorObserver) ObserveError(e *Error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Errors = append(c.Errors, e)
}

// ObservePeer implements Observer
func (c *CollectorObserver) ObservePeer(p *Peer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Peers = append(c.Peers, p)
}
