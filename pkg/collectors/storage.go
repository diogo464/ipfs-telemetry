package collectors

import (
	"context"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/corerepo"
)

var _ telemetry.Collector = (*storageCollector)(nil)

type storageCollector struct {
	node *core.IpfsNode
}

// Descriptor implements telemetry.Collector
func (*storageCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.StorageName,
	}
}

// Open implements telemetry.Collector
func (*storageCollector) Open() {}

func Storage(n *core.IpfsNode) telemetry.Collector {
	return &storageCollector{
		node: n,
	}
}

// Close implements Collector
func (*storageCollector) Close() {}

// Collect implements Collector
func (c *storageCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	stat, err := corerepo.RepoStat(ctx, c.node)
	if err != nil {
		return err
	}
	dp := &datapoint.Storage{
		Timestamp:    datapoint.NewTimestamp(),
		StorageUsed:  stat.RepoSize,
		StorageTotal: stat.SizeStat.StorageMax,
		NumObjects:   stat.NumObjects,
	}
	return datapoint.StorageSerialize(dp, stream)
}
