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
	"go.opentelemetry.io/otel/sdk/instrumentation"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

var ErrInvalidResponse = fmt.Errorf("invalid response")
var ErrNotUsingLibp2p = fmt.Errorf("not using libp2p")

var (
	clientStreamType_Metrics = 0
	clientStreamType_Events  = 1
)

type clientStreamKey struct {
	streamType int
	eventId    uint32
}

func newStreamKeyMetrics() clientStreamKey {
	return clientStreamKey{streamType: clientStreamType_Metrics}
}

func newStreamKeyEvent(eventId uint32) clientStreamKey {
	return clientStreamKey{streamType: clientStreamType_Metrics, eventId: eventId}
}

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
	sequenceNumbers map[clientStreamKey]uint32
}

func (s *ClientState) String() string {
	return fmt.Sprintf("Session: %s, SequenceNumbers: %v %v", s.session, s.sequenceNumbers)
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
			sequenceNumbers: make(map[clientStreamKey]uint32),
		}
	}

	if options.h != nil {
		conn, err := grpc.NewClient(
			"passthrough:",
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
		conn, err := grpc.NewClient(fmt.Sprintf("ipv4:%s", options.target), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		client.s.sequenceNumbers = make(map[clientStreamKey]uint32)
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
		property := Property{
			Scope: instrumentation.Scope{
				Name:    pbprop.GetScope().GetName(),
				Version: pbprop.GetScope().GetVersion(),
			},
			Name:        pbprop.GetName(),
			Description: pbprop.GetDescription(),
		}

		switch v := pbprop.GetValue().(type) {
		case *pb.Property_StringValue:
			property.Value = NewPropertyValueString(v.StringValue)
		case *pb.Property_IntegerValue:
			property.Value = NewPropertyValueInteger(v.IntegerValue)
		}

		properties = append(properties, property)
	}
	return properties, nil
}

func (c *Client) GetStream(ctx context.Context, key clientStreamKey) ([]stream.MessageBin, error) {
	segments, err := c.GetStreamSegments(ctx, key)
	if err != nil {
		return nil, err
	}

	messages := make([]stream.MessageBin, 0)
	for _, segment := range segments {
		msgs, err := stream.SegmentDecode(stream.ByteDecoder, segment)
		if err != nil {
			return nil, err
		}
		for _, msg := range msgs {
			messages = append(messages, stream.MessageBin(msg))
		}
	}

	return messages, nil
}

func (c *Client) GetStreamSegments(ctx context.Context, key clientStreamKey) ([]stream.Segment, error) {
	client, err := c.newGrpcClient()
	if err != nil {
		return nil, err
	}

	var srv grpc.ServerStreamingClient[pb.StreamSegment] = nil
	switch key.streamType {
	case clientStreamType_Metrics:
		srv, err = client.GetMetrics(ctx, &pb.GetMetricsRequest{})
		break
	case clientStreamType_Events:
		srv, err = client.GetEvents(ctx, &pb.GetEventsRequest{
			EventId:             key.eventId,
			SequenceNumberSince: c.s.sequenceNumbers[key],
		})
		break
	default:
		panic("unreachable")
	}

	if err != nil {
		return nil, err
	}

	segments := make([]stream.Segment, 0)
	for {
		pbseg, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		segment := pbSegmentToSegment(pbseg)
		segments = append(segments, segment)
		if uint32(segment.SeqN) > c.s.sequenceNumbers[key] {
			c.s.sequenceNumbers[key] = uint32(segment.SeqN)
		}
	}

	// If we received at least one segment, we need to increment the sequence number
	// to avoid requesting the same segment again next time we use this stream.
	if len(segments) > 0 {
		c.s.sequenceNumbers[key]++
	}

	return segments, nil
}

func (c *Client) GetMetrics(ctx context.Context) (Metrics, error) {
	messages, err := c.GetStream(ctx, newStreamKeyMetrics())
	if err != nil {
		return Metrics{}, err
	}

	metrics := make([]*v1.ResourceMetrics, len(messages))
	for i, msg := range messages {
		m := &v1.ResourceMetrics{}
		err = proto.Unmarshal(msg.Value, m)
		if err != nil {
			return Metrics{}, err
		}
		metrics[i] = m
	}

	return Metrics{
		OTLP: metrics,
	}, nil
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
		d, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		descriptors = append(descriptors, EventDescriptor{
			EventId: d.EventId,
			Scope: instrumentation.Scope{
				Name:    d.Scope.Name,
				Version: d.Scope.Version,
			},
			Name:        d.Name,
			Description: d.Description,
		})
	}

	return descriptors, nil
}

func (c *Client) GetEvents(ctx context.Context, eventId uint32) ([]Event, error) {
	messages, err := c.GetStream(ctx, newStreamKeyEvent(eventId))
	if err != nil {
		return nil, err
	}

	events := make([]Event, len(messages))
	for i, msg := range messages {
		events[i] = Event{
			Timestamp: msg.Timestamp,
			Data:      msg.Value,
		}
	}

	return events, nil
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
