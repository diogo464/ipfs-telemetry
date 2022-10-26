package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/diogo464/telemetry/pb"
	"github.com/diogo464/telemetry/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")
var ErrNotUsingLibp2p = fmt.Errorf("not using libp2p")

var _ io.Reader = (*propertyReader)(nil)

type propertyReader struct {
	b []byte
	c pb.Telemetry_GetPropertyClient
}

func newPropertyReader(c pb.Telemetry_GetPropertyClient) *propertyReader {
	return &propertyReader{c: c, b: []byte{}}
}

// Read implements io.Reader
func (r *propertyReader) Read(p []byte) (n int, err error) {
	for len(p) > 0 {
		if len(r.b) == 0 {
			seg, err := r.c.Recv()
			if err != nil {
				return n, err
			}
			r.b = seg.GetData()
		}
		n += copy(p, r.b)
		p = p[n:]
		r.b = r.b[n:]
	}
	return n, nil
}

type RawEvent struct {
	Timestamp time.Time
	// Json event data
	Data []byte
}

type RawSnapshot struct {
	Timestamp time.Time
	// Json snapshot data
	Data []byte
}

type GetMetricsResponse struct {
	// Next Sequence Number to request
	NextSeqN     int
	Observations []MetricsObservations
}

type GetEventResponse struct {
	// Next Sequence Number to request
	NextSeqN int
	Events   []RawEvent
}

type GetSnapshotResponse struct {
	// Next Sequence Number to request
	NextSeqN  int
	Snapshots []RawSnapshot
}

type AvailableEvent struct{ Name string }
type AvailableSnapshot struct{ Name string }
type AvailableProperty struct {
	Name     string
	Encoding Encoding
	Constant bool
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

func (c *Client) GetMetrics(ctx context.Context, since int) (*GetMetricsResponse, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetMetrics(ctx, &pb.GetMetricsRequest{})
	if err != nil {
		return nil, err
	}

	seqn := since
	observations := make([]MetricsObservations, 0)
	for {
		pbsegment, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		segment := pbSegmentToSegment(pbsegment)
		seqn = segment.SeqN + 1
		messages, err := StreamSegmentDecode(pbMetricsObservationsStreamDecoder, segment)
		if err != nil {
			return nil, err
		}

		for _, msg := range messages {
			observations = append(observations, MetricsObservations{
				Timestamp: msg.Timestamp,
				Metrics:   msg.Value.GetObservations(),
			})
		}
	}

	return &GetMetricsResponse{
		NextSeqN:     seqn,
		Observations: observations,
	}, nil
}

func (c *Client) GetAvailableEvents(ctx context.Context) ([]AvailableEvent, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetAvailableEvents(ctx, &pb.GetAvailableEventsRequest{})
	if err != nil {
		return nil, err
	}

	available := make([]AvailableEvent, 0)
	for {
		pbavail, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		available = append(available, AvailableEvent{
			Name: pbavail.Name,
		})
	}
	return available, nil
}

func (c *Client) GetEvent(ctx context.Context, name string, since int) (*GetEventResponse, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetEvent(ctx, &pb.GetEventRequest{
		Name: name,
		Seqn: uint32(since),
	})
	if err != nil {
		return nil, err
	}

	events := make([]RawEvent, 0)
	seqn := since
	for {
		pbsegment, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbsegment)
		seqn = segment.SeqN + 1
		messages, err := StreamSegmentDecode(ByteStreamDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			events = append(events, RawEvent{
				Timestamp: msg.Timestamp,
				Data:      msg.Value,
			})
		}
	}

	return &GetEventResponse{
		NextSeqN: seqn,
		Events:   events,
	}, nil
}

func (c *Client) GetAvailableSnapshots(ctx context.Context) ([]AvailableSnapshot, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetAvailableSnapshots(ctx, &pb.GetAvailableSnapshotsRequest{})
	if err != nil {
		return nil, err
	}

	available := make([]AvailableSnapshot, 0)
	for {
		pbavail, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		available = append(available, AvailableSnapshot{
			Name: pbavail.Name,
		})
	}
	return available, nil
}

func (c *Client) GetSnapshot(ctx context.Context, name string, since int) (*GetSnapshotResponse, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetSnapshot(ctx, &pb.GetSnapshotRequest{
		Name: name,
		Seqn: uint32(since),
	})
	if err != nil {
		return nil, err
	}

	events := make([]RawSnapshot, 0)
	seqn := since
	for {
		pbsegment, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbsegment)
		seqn = segment.SeqN + 1
		messages, err := StreamSegmentDecode(ByteStreamDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			events = append(events, RawSnapshot{
				Timestamp: msg.Timestamp,
				Data:      msg.Value,
			})
		}
	}

	return &GetSnapshotResponse{
		NextSeqN:  seqn,
		Snapshots: events,
	}, nil
}

func (c *Client) GetAvailableProperties(ctx context.Context) ([]PropertyDescriptor, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetAvailableProperties(context.Background(), &pb.GetAvailablePropertiesRequest{})
	if err != nil {
		return nil, err
	}

	properties := make([]PropertyDescriptor, 0)
	for {
		prop, err := response.Recv()
		if err != nil {
			if err == io.EOF || err == ctx.Err() {
				break
			}
			return nil, err
		}
		properties = append(properties, PropertyDescriptor{
			Name:     prop.GetName(),
			Encoding: Encoding(prop.GetEncoding()),
			Constant: prop.Constant,
		})
	}

	return properties, nil
}

func (c *Client) Property(ctx context.Context, property string) (io.Reader, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetProperty(ctx, &pb.GetPropertyRequest{
		Property: property,
	})
	if err != nil {
		return nil, err
	}

	return newPropertyReader(response), nil
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

func PropertyDecoded[T any](ctx context.Context, c *Client, name string, decoder PropertyDecoder[T]) (T, error) {
	r, e := c.Property(ctx, name)
	if e != nil {
		var v T
		return v, e
	}
	return decoder(r)
}

var pbMetricsObservationsStreamDecoder = func(b []byte) (*pb.MetricsObservations, error) {
	msg := &pb.MetricsObservations{}
	if err := proto.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func pbSegmentToSegment(s *pb.StreamSegment) StreamSegment {
	return StreamSegment{
		SeqN: int(s.Seqn),
		Data: s.GetData(),
	}
}
