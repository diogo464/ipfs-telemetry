package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*Connections)(nil)

const ConnectionsName = "connections"

type Stream struct {
	Protocol  string            `json:"protocol"`
	Opened    time.Time         `json:"opened"`
	Direction network.Direction `json:"direction"`
}

type Connection struct {
	ID      peer.ID             `json:"id"`
	Addr    multiaddr.Multiaddr `json:"addr"`
	Latency time.Duration       `json:"latency"`
	Opened  time.Time           `json:"opened"`
	Streams []Stream            `json:"streams"`
}

type Connections struct {
	Timestamp   time.Time    `json:"timestamp"`
	Connections []Connection `json:"connections"`
}

func (*Connections) sealed()                   {}
func (*Connections) GetName() string           { return ConnectionsName }
func (b *Connections) GetTimestamp() time.Time { return b.Timestamp }
func (b *Connections) GetSizeEstimate() uint32 {
	estimateStreamSize := 16 + estimateTimestampSize + 4
	estimateConnectionSize := estimatePeerIdSize + estimateMultiAddrSize + estimateDurationSize + estimateTimestampSize + estimateStreamSize
	return estimateTimestampSize + uint32(len(b.Connections))*uint32(estimateConnectionSize)
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

		streams := make([]Stream, 0, len(conn.GetStreams()))
		for _, pbstream := range conn.GetStreams() {
			streams = append(streams, Stream{
				Protocol:  pbstream.GetProtocol(),
				Opened:    pbstream.GetOpened().AsTime(),
				Direction: network.Direction(pbstream.GetDirection()),
			})
		}

		conns = append(conns, Connection{
			ID:      id,
			Addr:    addr,
			Latency: conn.GetLatency().AsDuration(),
			Opened:  conn.GetOpened().AsTime(),
			Streams: streams,
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
		pbstreams := make([]*pb.Connections_Stream, 0, len(conn.Streams))
		for _, stream := range conn.Streams {
			pbstreams = append(pbstreams, &pb.Connections_Stream{
				Protocol:  stream.Protocol,
				Opened:    timestamppb.New(stream.Opened),
				Direction: pb.Connections_Direction(stream.Direction),
			})
		}

		conns = append(conns, &pb.Connections_Connection{
			Peer:    conn.ID.String(),
			Addr:    conn.Addr.String(),
			Latency: durationpb.New(conn.Latency),
			Opened:  timestamppb.New(conn.Opened),
			Streams: pbstreams,
		})
	}

	return &pb.Connections{
		Timestamp:   timestamppb.New(c.Timestamp),
		Connections: conns,
	}
}
