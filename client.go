package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/diogo464/telemetry/pb"
	"github.com/diogo464/telemetry/utils"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gproto "google.golang.org/protobuf/proto"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")
var ErrNotUsingLibp2p = fmt.Errorf("not using libp2p")

type CProperty struct {
	Name        string
	Description string
	Value       PropertyValue
}

type Client struct {
	// Can be null if we are not connected using libp2p
	h host.Host
	p peer.ID

	c *grpc.ClientConn
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

func NewClient2(target string) (*Client, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		h: nil,
		p: "",
		c: conn,
	}, nil
}

func (c *Client) Close() {
	c.c.Close()
}

func (c *Client) GetSession(ctx context.Context) (Session, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return Session{}, err
	}

	response, err := client.GetSession(ctx, &pb.GetSessionRequest{})
	if err != nil {
		return Session{}, err
	}

	return ParseSession(response.GetUuid())
}

func (c *Client) GetMetrics(ctx context.Context, since int) ([]*mpb.ResourceMetrics, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{})
	if err != nil {
		return nil, err
	}

	mdatas := make([]*mpb.ResourceMetrics, 0)
	for {
		pbsegment, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		segment := pbSegmentToSegment(pbsegment)
		messages, err := StreamSegmentDecode(ByteStreamDecoder, segment)
		if err != nil {
			return nil, err
		}

		for _, msg := range messages {
			mdata := &mpb.ResourceMetrics{}
			if err := gproto.Unmarshal(msg.Value, mdata); err != nil {
				return nil, err
			}
			mdatas = append(mdatas, mdata)
		}
	}

	return mdatas, nil
}

func (c *Client) GetProperties(ctx context.Context) ([]CProperty, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetProperties(ctx, &pb.GetPropertiesRequest{})
	if err != nil {
		return nil, err
	}

	properties := make([]CProperty, 0)
	for {
		pbprop, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		properties = append(properties, propertyPbToClientProperty(pbprop))
	}
	return properties, nil
}

func (c *Client) GetCaptureDescriptors(ctx context.Context) ([]CaptureDescriptor, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetCaptureDescriptors(ctx, &pb.GetCaptureDescriptorsRequest{})
	if err != nil {
		return nil, err
	}

	descriptors := make([]CaptureDescriptor, 0)
	for {
		pbdesc, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, CaptureDescriptor{
			Name:        pbdesc.GetName(),
			Description: pbdesc.GetDescription(),
		})
	}

	return descriptors, nil
}

func (c *Client) GetCapture(ctx context.Context, name string, since int) ([]CaptureData, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetCapture(ctx, &pb.GetCaptureRequest{
		SequenceNumberSince: uint32(since),
		Name:                name,
	})
	if err != nil {
		return nil, err
	}

	datas := make([]CaptureData, 0)
	for {
		pbseg, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbseg)
		messages, err := StreamSegmentDecode(ByteStreamDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			datas = append(datas, CaptureData{
				SequenceNumber: segment.SeqN,
				Timestamp:      msg.Timestamp,
				Data:           msg.Value,
			})
		}
	}

	return datas, nil
}

func (c *Client) GetEventDescriptors(ctx context.Context) ([]EventDescriptor, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetEventDescriptors(ctx, &pb.GetEventDescriptorsRequest{})
	if err != nil {
		return nil, err
	}

	descriptors := make([]EventDescriptor, 0)
	for {
		pbdesc, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, EventDescriptor{
			Name:        pbdesc.GetName(),
			Description: pbdesc.GetDescription(),
		})
	}

	return descriptors, nil
}

func (c *Client) GetEvent(ctx context.Context, name string, since int) ([]EventData, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetEvent(ctx, &pb.GetEventRequest{
		SequenceNumberSince: uint32(since),
		Name:                name,
	})
	if err != nil {
		return nil, err
	}

	datas := make([]EventData, 0)
	for {
		pbseg, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbseg)
		messages, err := StreamSegmentDecode(ByteStreamDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			datas = append(datas, EventData{
				SequenceNumber: segment.SeqN,
				Timestamp:      msg.Timestamp,
				Data:           msg.Value,
			})
		}
	}

	return datas, nil
}

func (c *Client) Download(ctx context.Context, payload uint32) (uint32, error) {
	if c.h == nil {
		return 0, ErrNotUsingLibp2p
	}

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
	if c.h == nil {
		return 0, ErrNotUsingLibp2p
	}

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

func pbSegmentToSegment(s *pb.StreamSegment) StreamSegment {
	return StreamSegment{
		SeqN: int(s.GetSequenceNumber()),
		Data: s.GetData(),
	}
}
