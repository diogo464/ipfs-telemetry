package datapoint

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

const RelayReservationName = "relay_reservation"
const RelayConnectionName = "relay_connection"
const RelayCompleteName = "relay_complete"
const RelayStatsName = "relay_stats"

type RelayReservation struct {
	Timestamp time.Time `json:"timestamp"`
	Peer      peer.ID   `json:"peer"`
}

type RelayConnection struct {
	Timestamp time.Time `json:"timestamp"`
	Initiator peer.ID   `json:"initiator"`
	Target    peer.ID   `json:"target"`
}

type RelayComplete struct {
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	Initiator    peer.ID       `json:"initiator"`
	Target       peer.ID       `json:"target"`
	BytesRelayed uint64        `json:"bytes_relayed"`
}

type RelayStats struct {
	Timestamp         time.Time `json:"timestamp"`
	Reservations      uint32    `json:"reservations"`
	Connections       uint32    `json:"connections"`
	BytesRelayed      uint64    `json:"bytes_relayed"`
	ActiveConnections uint32    `json:"active_connections"`
}
