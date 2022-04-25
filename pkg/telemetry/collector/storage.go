package collector

import (
	"context"

	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corerepo"
)

var _ Collector = (*storageCollector)(nil)

type storageCollector struct {
	node *core.IpfsNode
}

func NewStorageCollector(n *core.IpfsNode) Collector {
	return &storageCollector{
		node: n,
	}
}

// Close implements Collector
func (*storageCollector) Close() {
}

// Collect implements Collector
func (c *storageCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	if stat, err := corerepo.RepoStat(ctx, c.node); err == nil {
		sink.Push(&datapoint.Storage{
			Timestamp:    datapoint.NewTimestamp(),
			StorageUsed:  stat.RepoSize,
			StorageTotal: stat.SizeStat.StorageMax,
			NumObjects:   stat.NumObjects,
		})
	}
}