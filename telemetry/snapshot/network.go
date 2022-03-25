package snapshot

import (
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

type NetworkSnapshot struct {
	TotalIn   uint64 `json:"totalin"`
	TotalOut  uint64 `json:"totalout"`
	RateIn    uint64 `json:"ratein"`
	RateOut   uint64 `json:"rateout"`
	NumConns  uint32 `json:"numconns"`
	LowWater  uint32 `json:"lowwater"`
	HighWater uint32 `json:"highwater"`
}

func NewNetworkSnapshot(ti uint64, to uint64, ri uint64, ro uint64, nc uint32, lw uint32, hw uint32) *Snapshot {
	return NewSnapshot("network", &NetworkSnapshot{
		TotalIn:   ti,
		TotalOut:  to,
		RateIn:    ri,
		RateOut:   ro,
		NumConns:  nc,
		LowWater:  lw,
		HighWater: hw,
	})
}

func NewNetworkSnapshotFromNode(n *core.IpfsNode) *Snapshot {
	reporter := n.Reporter
	totals := reporter.GetBandwidthTotals()
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	return NewNetworkSnapshot(uint64(totals.TotalIn), uint64(totals.TotalOut), uint64(totals.RateIn), uint64(totals.RateOut), uint32(info.ConnCount), uint32(info.LowWater), uint32(info.HighWater))
}
