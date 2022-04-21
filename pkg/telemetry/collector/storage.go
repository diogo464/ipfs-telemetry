package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corerepo"
)

type StorageOptions struct {
	Interval time.Duration
}

func RunStorageCollector(ctx context.Context, node *core.IpfsNode, sink snapshot.Sink, opts StorageOptions) {
	ticker := time.NewTicker(opts.Interval)

	for {
		select {
		case <-ticker.C:
			if stat, err := corerepo.RepoStat(ctx, node); err == nil {
				sink.Push(&snapshot.Storage{
					Timestamp:    snapshot.NewTimestamp(),
					StorageUsed:  stat.RepoSize,
					StorageTotal: stat.SizeStat.StorageMax,
					NumObjects:   stat.NumObjects,
				})
			}
		case <-ctx.Done():
		}
	}
}
