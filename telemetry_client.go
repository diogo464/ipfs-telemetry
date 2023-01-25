package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	"github.com/diogo464/telemetry/internal/utils"
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

type ClientOption = func(*clientOptions)

type clientOptions struct {
	// When h is not nil, dial using libp2p
	h host.Host
	p peer.ID

	// When h is nil, dial using grpc.Dial
	target string

	state *ClientState
}

type ClientState struct {
	session         Session
	sequenceNumbers map[StreamId]uint32
}

type Client struct {
	// Can be null if we are not connected using libp2p
	h host.Host
	p peer.ID
	s *ClientState
	c *grpc.ClientConn
}

func WithClientLibp2pDial(h host.Host, p peer.ID) ClientOption {
	return func(o *clientOptions) {
		o.h = h
		o.p = p
	}
}

func WithClientGrpcDial(target string) ClientOption {
	return func(o *clientOptions) {
		o.target = target
	}
}

func WithClientState(s *ClientState) ClientOption {
	return func(o *clientOptions) {
		o.state = s
	}
}

func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	options := new(clientOptions)
	for _, opt := range opts {
		opt(options)
	}

	client := new(Client)
	if options.state != nil {
		client.s = options.state
	} else {
		client.s = &ClientState{
			session:         Session{},
			sequenceNumbers: make(map[StreamId]uint32),
		}
	}

	if options.h != nil {
		conn, err := grpc.Dial(
			"",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				conn, err := gostream.Dial(ctx, options.h, options.p, ID_TELEMETRY)
				return conn, err
			}))

		if err != nil {
			return nil, err
		}

		client.h = options.h
		client.p = options.p
		client.c = conn
	} else {
		conn, err := grpc.Dial(options.target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		client.c = conn
	}

	sess, err := client.GetSession(ctx)
	if err != nil {
		return nil, err
	}

	if sess != client.s.session {
		client.s.session = sess
		client.s.sequenceNumbers = make(map[StreamId]uint32)
	}

	return client, nil
}

func (c *Client) Close() {
	c.c.Close()
}

func (c *Client) GetClientState() *ClientState {
	return c.s
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

func (c *Client) GetMetricDescriptors(ctx context.Context) ([]MetricDescriptor, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	response, err := client.GetMetricDescriptors(ctx, &pb.GetMetricDescriptorsRequest{})
	if err != nil {
		return nil, err
	}

	descriptors := make([]MetricDescriptor, 0)
	for {
		pbdesc, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, MetricDescriptor{
			Scope:       pbdesc.GetScope(),
			Name:        pbdesc.GetName(),
			Description: pbdesc.GetDescription(),
		})
	}

	return descriptors, nil
}

func (c *Client) GetMetrics(ctx context.Context) (Metrics, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return Metrics{}, err
	}

	response, err := client.GetStream(ctx, &pb.GetStreamRequest{
		StreamId:            uint32(METRICS_STREAM_ID),
		SequenceNumberSince: c.s.sequenceNumbers[METRICS_STREAM_ID],
	})
	if err != nil {
		return Metrics{}, err
	}

	mdatas := make([]*mpb.ResourceMetrics, 0)
	seqn := 0
	for {
		pbsegment, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Metrics{}, err
		}

		segment := pbSegmentToSegment(pbsegment)
		if segment.SeqN > seqn {
			seqn = segment.SeqN
		}
		messages, err := stream.SegmentDecode(stream.ByteDecoder, segment)
		if err != nil {
			return Metrics{}, err
		}

		for _, msg := range messages {
			mdata := &mpb.ResourceMetrics{}
			if err := gproto.Unmarshal(msg.Value, mdata); err != nil {
				return Metrics{}, err
			}
			mdatas = append(mdatas, mdata)
		}
	}

	c.s.sequenceNumbers[METRICS_STREAM_ID] = uint32(seqn)

	return Metrics{OTLP: mdatas}, nil
}

func (c *Client) GetPropertyDescriptors(ctx context.Context) ([]PropertyDescriptor, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetPropertyDescriptors(ctx, &pb.GetPropertyDescriptorsRequest{})
	if err != nil {
		return nil, err
	}

	descriptors := make([]PropertyDescriptor, 0)
	for {
		pbdesc, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, PropertyDescriptor{
			ID:          pbdesc.GetId(),
			Scope:       pbdesc.GetScope(),
			Name:        pbdesc.GetName(),
			Description: pbdesc.GetDescription(),
		})
	}
	return descriptors, nil
}

func (c *Client) GetProperties(ctx context.Context) ([]Property, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetProperties(ctx, &pb.GetPropertiesRequest{})
	if err != nil {
		return nil, err
	}

	properties := make([]Property, 0)
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
		descriptors = append(descriptors, captureDescriptorFromPb(pbdesc))
	}

	return descriptors, nil
}

func (c *Client) GetCapture(ctx context.Context, streamId StreamId) ([]Capture, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetStream(ctx, &pb.GetStreamRequest{
		StreamId:            uint32(streamId),
		SequenceNumberSince: uint32(c.s.sequenceNumbers[streamId]),
	})
	if err != nil {
		return nil, err
	}

	datas := make([]Capture, 0)
	for {
		pbseg, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbseg)
		messages, err := stream.SegmentDecode(stream.ByteDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			datas = append(datas, Capture{
				Timestamp: msg.Timestamp,
				Data:      msg.Value,
			})
		}

		if uint32(segment.SeqN) > c.s.sequenceNumbers[streamId] {
			c.s.sequenceNumbers[streamId] = uint32(segment.SeqN)
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
			StreamId:    StreamId(pbdesc.GetStreamId()),
			Scope:       pbdesc.GetScope(),
			Name:        pbdesc.GetName(),
			Description: pbdesc.GetDescription(),
		})
	}

	return descriptors, nil
}

func (c *Client) GetEvent(ctx context.Context, streamId StreamId) ([]Event, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	srv, err := client.GetStream(ctx, &pb.GetStreamRequest{
		StreamId:            uint32(streamId),
		SequenceNumberSince: uint32(c.s.sequenceNumbers[streamId]),
	})
	if err != nil {
		return nil, err
	}

	datas := make([]Event, 0)
	for {
		pbseg, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbseg)
		messages, err := stream.SegmentDecode(stream.ByteDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range messages {
			datas = append(datas, Event{
				Timestamp: msg.Timestamp,
				Data:      msg.Value,
			})
		}

		if uint32(segment.SeqN) > c.s.sequenceNumbers[streamId] {
			c.s.sequenceNumbers[streamId] = uint32(segment.SeqN)
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

func pbSegmentToSegment(s *pb.StreamSegment) stream.Segment {
	return stream.Segment{
		SeqN: int(s.GetSequenceNumber()),
		Data: s.GetData(),
	}
}
