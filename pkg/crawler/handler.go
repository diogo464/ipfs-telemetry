package crawler

import "github.com/libp2p/go-libp2p-core/peer"

type EventHandler interface {
	OnConnect(p peer.ID)
	OnFinish(p peer.ID, addrs []peer.AddrInfo)
	OnFail(p peer.ID, err error)
}

type NullEventHandler struct{}

func (NullEventHandler) OnConnect(p peer.ID)                       {}
func (NullEventHandler) OnFinish(p peer.ID, addrs []peer.AddrInfo) {}
func (NullEventHandler) OnFail(p peer.ID, err error)               {}
