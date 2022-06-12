package collectors

import (
	"context"

	"github.com/diogo464/telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corerepo"
)

var _ telemetry.Collector = (*storageCollector)(nil)

type storageCollector struct {
	node *core.IpfsNode
}

func Storage(n *core.IpfsNode) telemetry.Collector {
	return &storageCollector{
		node: n,
	}
}

// Name implements telemetry.Collector
func (*storageCollector) Name() string {
	return "Storage"
}

// Close implements Collector
func (*storageCollector) Close() {
}

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
