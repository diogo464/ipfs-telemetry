package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoutingTable struct {
	Timestamp time.Time   `json:"timestamp"`
	Buckets   [][]peer.ID `json:"buckets"`
}

func RoutingTableFromPB(in *pb.Snapshot_RoutingTable) (*RoutingTable, error) {
	buckets := make([][]peer.ID, 0, len(in.GetBuckets()))
	for _, b := range in.GetBuckets() {
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
		Timestamp: in.GetTimestamp().AsTime(),
		Buckets:   buckets,
	}, nil
}

func (r *RoutingTable) ToPB() *pb.Snapshot_RoutingTable {
	buckets := make([]*pb.Snapshot_RoutingTable_Bucket, 0, len(r.Buckets))
	for _, b := range r.Buckets {
		peers := make([]string, 0, len(b))
		for _, p := range b {
			peers = append(peers, p.Pretty())
		}
		buckets = append(buckets, &pb.Snapshot_RoutingTable_Bucket{
			Peers: peers,
		})
	}
	return &pb.Snapshot_RoutingTable{
		Timestamp: timestamppb.New(r.Timestamp),
		Buckets:   buckets,
	}
}

func RoutingTableArrayToPB(in []*RoutingTable) []*pb.Snapshot_RoutingTable {
	out := make([]*pb.Snapshot_RoutingTable, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}
