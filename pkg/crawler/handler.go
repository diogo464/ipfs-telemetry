package crawler

import "github.com/libp2p/go-libp2p-core/peer"

type EventHandler interface {
	OnConnect(p peer.ID) error
	OnFinish(p peer.ID, addrs []peer.AddrInfo) error
	OnFail(p peer.ID, err error) error
}

type NullEventHandler struct{}

func (NullEventHandler) OnConnect(p peer.ID) error                       { return nil }
func (NullEventHandler) OnFinish(p peer.ID, addrs []peer.AddrInfo) error { return nil }
func (NullEventHandler) OnFail(p peer.ID, err error) error               { return nil }
