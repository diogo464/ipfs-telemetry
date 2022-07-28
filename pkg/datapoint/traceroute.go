package datapoint

import (
	"time"

	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
)

const TraceRouteName = "traceroute"

type TraceRoute struct {
	Timestamp   time.Time
	Origin      peer.AddrInfo `json:"origin"`
	Destination peer.AddrInfo `json:"destination"`
	Provider    string        `json:"provider"`
	Output      []byte        `json:"output"`
}

func TraceRouteSerialize(in *TraceRoute, stream *telemetry.Stream) error {
	inpb := &pb.TraceRoute{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		Origin:      pbutils.AddrInfoToPB(&in.Origin),
		Destination: pbutils.AddrInfoToPB(&in.Destination),
		Provider:    in.Provider,
		Output:      in.Output,
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalTo(b)
		return err
	})
}

func TraceRouteDeserialize(in []byte) (*TraceRoute, error) {
	var inpb pb.TraceRoute
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	origin, err := pbutils.AddrInfoFromPB(inpb.GetOrigin())
	if err != nil {
		return nil, err
	}
	destination, err := pbutils.AddrInfoFromPB(inpb.GetDestination())
	if err != nil {
		return nil, err
	}
	return &TraceRoute{
		Timestamp:   pbutils.TimeFromPB(inpb.GetTimestamp()),
		Origin:      origin,
		Destination: destination,
		Provider:    inpb.GetProvider(),
		Output:      inpb.GetOutput(),
	}, nil
}
