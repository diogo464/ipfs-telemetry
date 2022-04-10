package crawler

import (
	"context"

	pb "git.d464.sh/adc/telemetry/pkg/proto/crawler"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	client pb.CrawlerClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		client: pb.NewCrawlerClient(conn),
	}
}

func (c *Client) Subscribe(ctx context.Context, sender chan<- peer.ID) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(sender)

	stream, err := c.client.Subscribe(ctx, &emptypb.Empty{})
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
