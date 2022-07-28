package datapoint

import (
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
)

const StorageName = "storage"

type Storage struct {
	Timestamp    time.Time `json:"timestamp"`
	StorageUsed  uint64    `json:"storage_used"`
	StorageTotal uint64    `json:"storage_total"`
	NumObjects   uint64    `json:"num_objects"`
}

func StorageSerialize(in *Storage, stream *telemetry.Stream) error {
	inpb := &pb.Storage{
		Timestamp:    pbutils.TimeToPB(&in.Timestamp),
		StorageUsed:  in.StorageUsed,
		StorageTotal: in.StorageTotal,
		NumObjects:   in.NumObjects,
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func StorageDeserialize(in []byte) (*Storage, error) {
	outpb := &pb.Storage{}
	err := outpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Timestamp:    pbutils.TimeFromPB(outpb.GetTimestamp()),
		StorageUsed:  outpb.GetStorageUsed(),
		StorageTotal: outpb.GetStorageTotal(),
		NumObjects:   outpb.GetNumObjects(),
	}, nil
}
