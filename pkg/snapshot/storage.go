package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Storage struct {
	Timestamp    time.Time `json:"timestamp"`
	StorageUsed  uint64    `json:"storage_used"`
	StorageTotal uint64    `json:"storage_total"`
	NumObjects   uint64    `json:"num_objects"`
}

func (*Storage) sealed() {}

func StorageFromPB(in *pb.Storage) (*Storage, error) {
	return &Storage{
		StorageUsed:  in.GetStorageUsed(),
		StorageTotal: in.GetStorageTotal(),
		NumObjects:   in.GetNumObjects(),
	}, nil
}

func (s *Storage) ToPB() *pb.Storage {
	return &pb.Storage{
		Timestamp:    timestamppb.New(s.Timestamp),
		StorageUsed:  s.StorageUsed,
		StorageTotal: s.StorageTotal,
		NumObjects:   s.NumObjects,
	}
}
