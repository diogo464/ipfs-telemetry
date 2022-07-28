package datapoint

import (
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
)

const RoutingTableName = "routingtable"

type RoutingTable struct {
	Timestamp time.Time   `json:"timestamp"`
	Buckets   [][]peer.ID `json:"buckets"`
}

func RoutingTableSerialize(in *RoutingTable, stream *telemetry.Stream) error {
	buckets := make([]*pb.RoutingTable_Bucket, 0, len(in.Buckets))
	for _, b := range in.Buckets {
		peers := make([]string, 0, len(b))
		for _, p := range b {
			peers = append(peers, p.Pretty())
		}
		buckets = append(buckets, &pb.RoutingTable_Bucket{
			Peers: peers,
		})
	}
	inpb := &pb.RoutingTable{
		Timestamp: pbutils.TimeToPB(&in.Timestamp),
		Buckets:   buckets,
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func RoutingTableDeserialize(in []byte) (*RoutingTable, error) {
	var inpb pb.RoutingTable

	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	buckets := make([][]peer.ID, 0, len(inpb.GetBuckets()))
	for _, b := range inpb.GetBuckets() {
		bucket := make([]peer.ID, 0, len(b.GetPeers()))
		for _, p := range b.GetPeers() {
			pid, err := peer.Decode(p)
			if err != nil {
				return nil, err
			}
			bucket = append(bucket, pid)
		}
		buckets = append(buckets, bucket)
	}

	return &RoutingTable{
		Timestamp: pbutils.TimeFromPB(inpb.GetTimestamp()),
		Buckets:   buckets,
	}, nil
}
