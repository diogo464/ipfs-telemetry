package monitor

import (
	"context"

	"github.com/diogo464/telemetry/monitor/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		c: pb.NewMonitorClient(conn),
	}
}

type Client struct {
	c pb.MonitorClient
}

func (c *Client) Discover(ctx context.Context, p peer.ID) error {
	_, err := c.c.Discover(ctx, &pb.DiscoverRequest{
		Peer: p.Pretty(),
	})
	return err
}
