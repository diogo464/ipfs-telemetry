package datapoint

import (
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

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

func ConnectionsSerialize(in *Connections, stream *telemetry.Stream) error {
	conns := make([]*pb.Connections_Connection, 0, len(in.Connections))
	for _, conn := range in.Connections {
		pbstreams := make([]*pb.Connections_Stream, 0, len(conn.Streams))
		for _, stream := range conn.Streams {
			pbstreams = append(pbstreams, &pb.Connections_Stream{
				Protocol:  stream.Protocol,
				Opened:    pbutils.TimeToPB(&stream.Opened),
				Direction: pb.Connections_Direction(stream.Direction),
			})
		}

		conns = append(conns, &pb.Connections_Connection{
			Peer:    conn.ID.String(),
			Addr:    conn.Addr.String(),
			Latency: pbutils.DurationToPB(&conn.Latency),
			Opened:  pbutils.TimeToPB(&conn.Opened),
			Streams: pbstreams,
		})
	}

	conns_pb := &pb.Connections{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		Connections: conns,
	}

	return stream.AllocAndWrite(conns_pb.Size(), func(b []byte) error {
		_, err := conns_pb.MarshalToSizedBuffer(b)
		return err
	})
}

func ConnectionsDeserialize(in []byte) (*Connections, error) {
	var pbin pb.Connections
	err := pbin.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	conns := make([]Connection, 0, len(pbin.GetConnections()))
	for _, conn := range pbin.GetConnections() {
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
				Opened:    pbutils.TimeFromPB(pbstream.GetOpened()),
				Direction: network.Direction(pbstream.GetDirection()),
			})
		}

		conns = append(conns, Connection{
			ID:      id,
			Addr:    addr,
			Latency: pbutils.DurationFromPB(conn.GetLatency()),
			Opened:  pbutils.TimeFromPB(conn.GetOpened()),
			Streams: streams,
		})
	}

	return &Connections{
		Timestamp:   pbutils.TimeFromPB(pbin.Timestamp),
		Connections: conns,
	}, nil
}
