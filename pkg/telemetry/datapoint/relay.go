package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*RelayReservation)(nil)
var _ Datapoint = (*RelayConnection)(nil)
var _ Datapoint = (*RelayComplete)(nil)
var _ Datapoint = (*RelayStats)(nil)

const RelayReservationName = "relay_reservation"
const RelayConnectionName = "relay_connection"
const RelayCompleteName = "relay_complete"
const RelayStatsName = "relay_stats"

type RelayReservation struct {
	Timestamp time.Time `json:"timestamp"`
	Peer      peer.ID   `json:"peer"`
}

func (*RelayReservation) sealed()                   {}
func (*RelayReservation) GetName() string           { return RelayReservationName }
func (r *RelayReservation) GetTimestamp() time.Time { return r.Timestamp }
func (r *RelayReservation) GetSizeEstimate() uint32 {
	return estimateTimestampSize + estimatePeerIdSize
}
func (r *RelayReservation) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_RelayReservation{
			RelayReservation: RelayReservationToPB(r),
		},
	}
}

func RelayReservationFromPB(in *pb.RelayReservation) (*RelayReservation, error) {
	p, err := peer.Decode(in.GetPeer())
	if err != nil {
		return nil, err
	}
	return &RelayReservation{
		Timestamp: in.GetTimestamp().AsTime(),
		Peer:      p,
	}, nil
}

func RelayReservationToPB(r *RelayReservation) *pb.RelayReservation {
	return &pb.RelayReservation{
		Timestamp: timestamppb.New(r.Timestamp),
		Peer:      r.Peer.String(),
	}
}

type RelayConnection struct {
	Timestamp time.Time `json:"timestamp"`
	Initiator peer.ID   `json:"initiator"`
	Target    peer.ID   `json:"target"`
}

func (*RelayConnection) sealed()                   {}
func (*RelayConnection) GetName() string           { return RelayConnectionName }
func (r *RelayConnection) GetTimestamp() time.Time { return r.Timestamp }
func (r *RelayConnection) GetSizeEstimate() uint32 {
	return estimateTimestampSize + estimatePeerIdSize*2
}
func (r *RelayConnection) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_RelayConnection{
			RelayConnection: RelayConnectionToPB(r),
		},
	}
}

func RelayConnectionFromPB(in *pb.RelayConnection) (*RelayConnection, error) {
	initiator, err := peer.Decode(in.GetInitiator())
	if err != nil {
		return nil, err
	}
	target, err := peer.Decode(in.GetTarget())
	if err != nil {
		return nil, err
	}
	return &RelayConnection{
		Timestamp: in.GetTimestamp().AsTime(),
		Initiator: initiator,
		Target:    target,
	}, nil
}

func RelayConnectionToPB(n *RelayConnection) *pb.RelayConnection {
	return &pb.RelayConnection{
		Timestamp: timestamppb.New(n.Timestamp),
		Initiator: n.Initiator.String(),
		Target:    n.Target.String(),
	}
}

type RelayComplete struct {
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	Initiator    peer.ID       `json:"initiator"`
	Target       peer.ID       `json:"target"`
	BytesRelayed uint64        `json:"bytes_relayed"`
}

func (*RelayComplete) sealed()                   {}
func (*RelayComplete) GetName() string           { return RelayCompleteName }
func (r *RelayComplete) GetTimestamp() time.Time { return r.Timestamp }
func (r *RelayComplete) GetSizeEstimate() uint32 {
	return estimateTimestampSize + estimateDurationSize + estimatePeerIdSize*2 + 8
}
func (r *RelayComplete) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_RelayComplete{
			RelayComplete: RelayCompleteToPB(r),
		},
	}
}

func RelayCompleteFromPB(in *pb.RelayComplete) (*RelayComplete, error) {
	initiator, err := peer.Decode(in.GetInitiator())
	if err != nil {
		return nil, err
	}
	target, err := peer.Decode(in.GetTarget())
	if err != nil {
		return nil, err
	}
	return &RelayComplete{
		Timestamp:    in.GetTimestamp().AsTime(),
		Duration:     in.GetDuration().AsDuration(),
		Initiator:    initiator,
		Target:       target,
		BytesRelayed: in.GetBytesRelayed(),
	}, nil
}

func RelayCompleteToPB(r *RelayComplete) *pb.RelayComplete {
	return &pb.RelayComplete{
		Timestamp:    timestamppb.New(r.Timestamp),
		Duration:     durationpb.New(r.Duration),
		Initiator:    r.Initiator.String(),
		Target:       r.Target.String(),
		BytesRelayed: r.BytesRelayed,
	}
}

type RelayStats struct {
	Timestamp         time.Time `json:"timestamp"`
	Reservations      uint32    `json:"reservations"`
	Connections       uint32    `json:"connections"`
	BytesRelayed      uint64    `json:"bytes_relayed"`
	ActiveConnections uint32    `json:"active_connections"`
}

func (*RelayStats) sealed()                   {}
func (*RelayStats) GetName() string           { return RelayStatsName }
func (r *RelayStats) GetTimestamp() time.Time { return r.Timestamp }
func (r *RelayStats) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 4 + 4 + 8 + 4
}
func (r *RelayStats) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_RelayStats{
			RelayStats: RelayStatsToPB(r),
		},
	}
}

func RelayStatsFromPB(in *pb.RelayStats) (*RelayStats, error) {
	return &RelayStats{
		Timestamp:         in.GetTimestamp().AsTime(),
		Reservations:      in.GetReservations(),
		Connections:       in.GetConnections(),
		BytesRelayed:      in.GetBytesRelayed(),
		ActiveConnections: in.GetActiveConnections(),
	}, nil
}

func RelayStatsToPB(r *RelayStats) *pb.RelayStats {
	return &pb.RelayStats{
		Timestamp:         timestamppb.New(r.Timestamp),
		Reservations:      r.Reservations,
		Connections:       r.Connections,
		BytesRelayed:      r.BytesRelayed,
		ActiveConnections: r.ActiveConnections,
	}
}
