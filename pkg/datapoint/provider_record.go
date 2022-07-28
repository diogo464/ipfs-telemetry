package datapoint

import (
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
)

const ProviderRecordsName = "provider_records"

type ProviderRecord struct {
	Key         []byte    `json:"key"`
	Peer        peer.ID   `json:"peer"`
	LastRefresh time.Time `json:"last_refresh"`
}

func ProviderRecordSerialize(in *ProviderRecord, stream *telemetry.Stream) error {
	pid, err := in.Peer.MarshalBinary()
	if err != nil {
		return err
	}

	dp := &pb.ProviderRecord{
		Key:         in.Key,
		Peer:        pid,
		LastRefresh: pbutils.TimeToPB(&in.LastRefresh),
	}

	return stream.AllocAndWrite(dp.Size(), func(buf []byte) error {
		_, err := dp.MarshalToSizedBuffer(buf)
		return err
	})
}

func ProviderRecordDeserialize(in []byte) (*ProviderRecord, error) {
	var inpb pb.ProviderRecord
	if err := inpb.Unmarshal(in); err != nil {
		return nil, err
	}

	pid, err := peer.IDFromBytes(inpb.GetPeer())
	if err != nil {
		return nil, err
	}

	return &ProviderRecord{
		Key:         inpb.GetKey(),
		Peer:        pid,
		LastRefresh: pbutils.TimeFromPB(inpb.GetLastRefresh()),
	}, nil
}
