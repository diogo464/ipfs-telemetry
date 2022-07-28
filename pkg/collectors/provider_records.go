package collectors

import (
	"context"
	"io"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/ipfs_telemetry/pkg/rle"
	"github.com/diogo464/telemetry"
	"github.com/ipfs/kubo/core"
)

var _ telemetry.Property = (*providerRecordsProperty)(nil)

type providerRecordsProperty struct {
	node *core.IpfsNode
}

func ProviderRecordsProperty(node *core.IpfsNode) telemetry.Property {
	return &providerRecordsProperty{
		node: node,
	}
}

// Collect implements telemetry.Property
func (p *providerRecordsProperty) Collect(ctx context.Context, writer io.Writer) error {
	c, err := p.node.DHT.WAN.ProviderStore().GetProviderRecords(ctx)
	if err != nil {
		return err
	}

	for in := range c {
		pid, err := in.Peer.MarshalBinary()
		if err != nil {
			return err
		}

		dp := &pb.ProviderRecord{
			Key:         in.Key,
			Peer:        pid,
			LastRefresh: pbutils.TimeToPB(&in.LastRefresh),
		}

		marshaled, err := dp.Marshal()
		if err != nil {
			return err
		}

		if err := rle.Write(writer, marshaled); err != nil {
			return err
		}
	}
	return nil
}

// Descriptor implements telemetry.Property
func (*providerRecordsProperty) Descriptor() telemetry.PropertyDescriptor {
	return telemetry.PropertyDescriptor{
		Name: datapoint.ProviderRecordsName,
	}
}
