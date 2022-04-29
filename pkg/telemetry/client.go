package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")

type DatapointStreamItem struct {
	NextSeqN   uint64
	Datapoints []datapoint.Datapoint
}

type Client struct {
	h host.Host
	p peer.ID
	c *grpc.ClientConn
}

func Connect(ctx context.Context, h host.Host, p peer.ID) (*Client, error) {
	stream, err := gostream.Dial(ctx, h, p, ID_TELEMETRY)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return stream, err
		}))
	return &Client{
		h: h,
		p: p,
		c: conn,
	}, nil
}

func NewClient(h host.Host, p peer.ID) (*Client, error) {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := gostream.Dial(ctx, h, p, ID_TELEMETRY)
			return conn, err
		}))

	if err != nil {
		return nil, err
	}

	return &Client{
		h: h,
		p: p,
		c: conn,
	}, nil
}

func (c *Client) Close() {
	c.c.Close()
}

func (c *Client) SessionInfo(ctx context.Context) (*SessionInfo, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSessionInfo(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	session, err := ParseSession(response.GetSession())
	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		Session:  session,
		BootTime: response.GetBootTime().AsTime(),
	}, nil
}

func (c *Client) Datapoints(ctx context.Context, since uint64, css chan<- DatapointStreamItem) error {
	client, err := c.newGrpcClient()
	if err != nil {
		return err
	}

	stream, err := client.GetDatapoints(ctx, &pb.GetDatapointsRequest{
		Since: since,
	})
	if err != nil {
		return err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		datapointspb := response.GetDatapoints()
		datapoints, err := datapoint.FromArrayPB(datapointspb)
		if err != nil {
			return err
		}

		css <- DatapointStreamItem{
			NextSeqN:   response.GetNext(),
			Datapoints: datapoints,
		}
	}

	return nil
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

func (c *Client) Bandwidth(ctx context.Context, payload uint32) (Bandwidth, error) {
	download, err := c.Download(ctx, payload)
	if err != nil {
		return Bandwidth{}, err
	}
	upload, err := c.Upload(ctx, payload)
	if err != nil {
		return Bandwidth{}, err
	}
	return Bandwidth{
		UploadRate:   upload,
		DownloadRate: download,
	}, nil
}

func (c *Client) newGrpcClient() (pb.TelemetryClient, error) {
	return pb.NewTelemetryClient(c.c), nil
}
