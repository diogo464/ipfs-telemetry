package measurements

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

var relay Relay = nil

type Relay interface {
	Reservation(p peer.ID)
	ConnectionOpen(initiator peer.ID, target peer.ID)
	ConnectionClose(initiator peer.ID, target peer.ID, duration time.Duration, relayed uint64)
}

func RelayRegister(r Relay) {
	if relay != nil {
		panic("should not happen")
	}
	relay = r
}

func WithRelay(fn func(Relay)) {
	if relay != nil {
		fn(relay)
	}
}
