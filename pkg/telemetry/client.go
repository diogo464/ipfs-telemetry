package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")

type Client struct {
	// session uuid
	s uuid.NullUUID
	// sequence number for snapshots
	n uint64

	h host.Host
	p peer.ID
	c *grpc.ClientConn
}

func NewClient(h host.Host, p peer.ID) (*Client, error) {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			fmt.Println("Grpc Dialing Peer")
			conn, err := gostream.Dial(ctx, h, p, ID_TELEMETRY)
			fmt.Println("Dial Complete")
			return conn, err
		}))

	if err != nil {
		return nil, err
	}

	return &Client{
		s: uuid.NullUUID{},
		n: 0,
		h: h,
		p: p,
		c: conn,
	}, nil
}

func (c *Client) Close() {
	c.c.Close()
}

func (c *Client) Session() uuid.NullUUID {
	return c.s
}

func (c *Client) Snapshots(ctx context.Context) ([]snapshot.Snapshot, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSnapshots(ctx, &pb.GetSnapshotsRequest{
		Session: c.s.UUID.String(),
		Since:   c.n,
	})
	if err != nil {
		return nil, err
	}

	session, err := uuid.Parse(response.Session)
	if err != nil {
		return nil, err
	}

	snapshots := response.GetSnapshots()
	converted := make([]snapshot.Snapshot, len(snapshots))
	for i, s := range snapshots {
		v, err := snapshot.FromPB(s)
		if err != nil {
			return nil, err
		}
		converted[i] = v
	}

	c.s.UUID = session
	c.s.Valid = true
	c.n = response.Next

	return converted, nil
}

func (c *Client) SystemInfo(ctx context.Context) (*SystemInfo, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSystemInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		OS:     response.Os,
		Arch:   response.Arch,
		NumCPU: response.Numcpu,
	}, nil
}

func (c *Client) Download(ctx context.Context, payload uint32) (uint32, error) {
	stream, err := c.h.NewStream(ctx, c.p, ID_DOWNLOAD)
	if err != nil {
		return 0, err
	}

	if err := utils.WriteU32(stream, payload); err != nil {
		return 0, err
	}

	if _, err := io.Copy(stream, io.LimitReader(utils.NullReader{}, int64(payload))); err != nil {
		return 0, err
	}

	rate, err := utils.ReadU32(stream)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

func (c *Client) Upload(ctx context.Context, payload uint32) (uint32, error) {
	stream, err := c.h.NewStream(ctx, c.p, ID_UPLOAD)
	if err != nil {
		return 0, err
	}

	if err := utils.WriteU32(stream, payload); err != nil {
		return 0, err
	}

	if _, err := io.Copy(io.Discard, io.LimitReader(stream, int64(payload))); err != nil {
		return 0, err
	}

	rate, err := utils.ReadU32(stream)
	if err != nil {
		return 0, err
	}

	return rate, nil
}

func (c *Client) newGrpcClient() (pb.TelemetryClient, error) {
	return pb.NewTelemetryClient(c.c), nil
}
