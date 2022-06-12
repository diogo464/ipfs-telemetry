package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/gogo/protobuf/types"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")

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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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

	response, err := client.GetSessionInfo(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}

	session, err := ParseSession(response.GetSession())
	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		Session:  session,
		BootTime: pbutils.TimeFromPB(response.GetBootTime()),
	}, nil
}

func (c *Client) AvailableStreams(ctx context.Context) ([]string, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetAvailableStreams(ctx, &pb.GetAvailableStreamsRequest{})
	if err != nil {
		return nil, err
	}

	return resp.GetStreams(), nil
}

func (c *Client) Stream(ctx context.Context, since uint32, stream string) ([]StreamSegment, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetStream(ctx, &pb.GetStreamRequest{
		Stream: stream,
		Seqn:   since,
	})
	if err != nil {
		return nil, err
	}

	segments := make([]StreamSegment, 0)

	for {
		seg, err := srv.Recv()
		if err == io.EOF {
			break
		}

		segments = append(segments, StreamSegment{
			SeqN: int(seg.GetSeqn()),
			Data: seg.GetData(),
		})
	}

	return segments, nil
}

func (c *Client) SystemInfo(ctx context.Context) (*SystemInfo, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSystemInfo(ctx, &types.Empty{})
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

func (c *Client) ProviderRecords(ctx context.Context) (<-chan ProviderRecord, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	stream, err := client.GetProviderRecords(ctx, &pb.GetProviderRecordsRequest{})
	if err != nil {
		return nil, err
	}

	crecords := make(chan ProviderRecord)
	go func() {
		defer close(crecords)

		for {
			pbrecord, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			pid, err := peer.IDFromBytes(pbrecord.Peer)
			if err != nil {
				return
			}

			crecords <- ProviderRecord{
				Key:         pbrecord.Key,
				Peer:        pid,
				LastRefresh: pbutils.TimeFromPB(pbrecord.GetLastRefresh()),
			}
		}
	}()

	return crecords, nil
}

func (c *Client) Debug(ctx context.Context) (*Debug, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	pbdbg, err := client.GetDebug(ctx, &types.Empty{})
	if err != nil {
		return nil, err
	}

	streams := make([]DebugStream, 0, len(pbdbg.GetStreams()))
	for _, pbs := range pbdbg.GetStreams() {
		streams = append(streams, DebugStream{
			Name:      pbs.Name,
			UsedSize:  pbs.Used,
			TotalSize: pbs.Total,
		})
	}

	return &Debug{
		Streams: streams,
	}, nil
}

func (c *Client) newGrpcClient() (pb.TelemetryClient, error) {
	return pb.NewTelemetryClient(c.c), nil
}
