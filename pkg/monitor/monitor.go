package monitor

import (
	"context"

	"git.d464.sh/adc/telemetry/pkg/monitor/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
)

type Client interface {
	Discover(ctx context.Context, p peer.ID) error
}

func NewClient(conn grpc.ClientConnInterface) Client {
	return &client{
		c: pb.NewMonitorClient(conn),
	}
}

type client struct {
	c pb.MonitorClient
}

func (c *client) Discover(ctx context.Context, p peer.ID) error {
	_, err := c.c.Discover(ctx, &pb.DiscoverRequest{
		Peer: p.Pretty(),
	})
	return err
}
