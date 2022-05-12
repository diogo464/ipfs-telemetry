package probe

import (
	"context"
	"fmt"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/probe"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProbeResult struct {
	RequestStart    time.Time
	RequestDuration time.Duration
	Peer            peer.ID
	Error           error
}

type Client struct {
	c pb.ProbeClient
}

func NewClient(c *grpc.ClientConn) *Client {
	return &Client{
		c: pb.NewProbeClient(c),
	}
}

func (c *Client) GetName(ctx context.Context) (string, error) {
	resp, err := c.c.ProbeGetName(ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return resp.GetName(), nil
}

func (c *Client) ProbeSetCids(ctx context.Context, cids []cid.Cid) error {
	cidsStr := make([]string, 0, len(cids))
	for _, c := range cids {
		cidsStr = append(cidsStr, c.String())
	}
	req := &pb.ProbeSetCidsRequest{
		Cids: cidsStr,
	}
	_, err := c.c.ProbeSetCids(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ProbeResults(ctx context.Context, out chan<- *ProbeResult) error {
	stream, err := c.c.ProbeResults(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	for {
		pbresult, err := stream.Recv()
		if err != nil {
			return err
		}

		pid, err := peer.Decode(pbresult.GetPeer())
		if err != nil {
			return err
		}

		var resultErr error = nil
		if pbresult.GetError() != "" {
			resultErr = fmt.Errorf(pbresult.GetError())
		}

		result := &ProbeResult{
			RequestStart:    pbresult.RequestStart.AsTime(),
			RequestDuration: pbresult.GetRequestDuration().AsDuration(),
			Peer:            pid,
			Error:           resultErr,
		}

		out <- result
	}
}
