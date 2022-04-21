package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*RoutingTable)(nil)

const RoutingTableName = "routingtable"

type RoutingTable struct {
	Timestamp time.Time   `json:"timestamp"`
	Buckets   [][]peer.ID `json:"buckets"`
}

func (*RoutingTable) sealed()                   {}
func (*RoutingTable) GetName() string           { return RoutingTableName }
func (r *RoutingTable) GetTimestamp() time.Time { return r.Timestamp }
func (r *RoutingTable) GetSizeEstimate() uint32 {
	var totalPeers uint32 = 0
	for _, b := range r.Buckets {
		totalPeers += uint32(len(b))
	}
	return estimateTimestampSize + totalPeers*estimatePeerIdSize
}
func (r *RoutingTable) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_RoutingTable{
			RoutingTable: RoutingTableToPB(r),
		},
	}
}

func RoutingTableFromPB(in *pb.RoutingTable) (*RoutingTable, error) {
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

func RoutingTableToPB(r *RoutingTable) *pb.RoutingTable {
	buckets := make([]*pb.RoutingTable_Bucket, 0, len(r.Buckets))
	for _, b := range r.Buckets {
		peers := make([]string, 0, len(b))
		for _, p := range b {
			peers = append(peers, p.Pretty())
		}
		buckets = append(buckets, &pb.RoutingTable_Bucket{
			Peers: peers,
		})
	}
	return &pb.RoutingTable{
		Timestamp: timestamppb.New(r.Timestamp),
		Buckets:   buckets,
	}
}

func RoutingTableArrayToPB(in []*RoutingTable) []*pb.RoutingTable {
	out := make([]*pb.RoutingTable, 0, len(in))
	for _, p := range in {
		out = append(out, RoutingTableToPB(p))
	}
	return out
}
