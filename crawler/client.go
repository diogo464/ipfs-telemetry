package crawler

import (
	"context"

	"github.com/diogo464/telemetry/crawler/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

type Client struct {
	client pb.CrawlerClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		client: pb.NewCrawlerClient(conn),
	}
}

func (c *Client) Subscribe(ctx context.Context, sender chan<- peer.ID) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(sender)

	stream, err := c.client.Subscribe(ctx, &pb.SubscribeRequest{})
	if err != nil {
		return err
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}

		item, err := stream.Recv()
		if err != nil {
			return err
		}

		p, err := peer.Decode(item.GetPeerId())
		if err != nil {
			return err
		}

		sender <- p
	}

	return nil
}
