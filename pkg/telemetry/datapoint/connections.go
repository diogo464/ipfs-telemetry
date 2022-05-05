package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*Connections)(nil)

const ConnectionsName = "connections"

type Connection struct {
	ID      peer.ID             `json:"id"`
	Addr    multiaddr.Multiaddr `json:"addr"`
	Latency time.Duration       `json:"latency"`
}

type Connections struct {
	Timestamp   time.Time    `json:"timestamp"`
	Connections []Connection `json:"connections"`
}

func (*Connections) sealed()                   {}
func (*Connections) GetName() string           { return ConnectionsName }
func (b *Connections) GetTimestamp() time.Time { return b.Timestamp }
func (b *Connections) GetSizeEstimate() uint32 {
	return estimateTimestampSize + uint32(len(b.Connections))*(estimatePeerIdSize+estimateMultiAddrSize+estimateDurationSize)
}
func (c *Connections) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Connections{
			Connections: ConnectionsToPB(c),
		},
	}
}

func ConnectionsFromPB(in *pb.Connections) (*Connections, error) {
	conns := make([]Connection, 0, len(in.GetConnections()))
	for _, conn := range in.GetConnections() {
		id, err := peer.Decode(conn.GetPeer())
		if err != nil {
			return nil, err
		}

		addr, err := multiaddr.NewMultiaddr(conn.GetAddr())
		if err != nil {
			return nil, err
		}

		conns = append(conns, Connection{
			ID:      id,
			Addr:    addr,
			Latency: conn.GetLatency().AsDuration(),
		})
	}

	return &Connections{
		Timestamp:   in.Timestamp.AsTime(),
		Connections: conns,
	}, nil
}

func ConnectionsToPB(c *Connections) *pb.Connections {
	conns := make([]*pb.Connections_Connection, 0, len(c.Connections))
	for _, conn := range c.Connections {
		conns = append(conns, &pb.Connections_Connection{
			Peer:    conn.ID.String(),
			Addr:    conn.Addr.String(),
			Latency: durationpb.New(conn.Latency),
		})
	}

	return &pb.Connections{
		Timestamp:   timestamppb.New(c.Timestamp),
		Connections: conns,
	}
}
