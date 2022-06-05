package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
)

var _ Datapoint = (*Storage)(nil)

const StorageName = "storage"

type Storage struct {
	Timestamp    time.Time `json:"timestamp"`
	StorageUsed  uint64    `json:"storage_used"`
	StorageTotal uint64    `json:"storage_total"`
	NumObjects   uint64    `json:"num_objects"`
}

func (*Storage) sealed()                   {}
func (*Storage) GetName() string           { return StorageName }
func (s *Storage) GetTimestamp() time.Time { return s.Timestamp }
func (s *Storage) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 3*8
}
func (s *Storage) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Storage{
			Storage: StorageToPB(s),
		},
	}
}

func StorageFromPB(in *pb.Storage) (*Storage, error) {
	return &Storage{
		Timestamp:    pbutils.TimeFromPB(in.GetTimestamp()),
		StorageUsed:  in.GetStorageUsed(),
		StorageTotal: in.GetStorageTotal(),
		NumObjects:   in.GetNumObjects(),
	}, nil
}

func StorageToPB(s *Storage) *pb.Storage {
	return &pb.Storage{
		Timestamp:    pbutils.TimeToPB(&s.Timestamp),
		StorageUsed:  s.StorageUsed,
		StorageTotal: s.StorageTotal,
		NumObjects:   s.NumObjects,
	}
}
