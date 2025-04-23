package telemetry

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	"github.com/diogo464/telemetry/metrics"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	TAG_UPLOAD     = "upload"
	TAG_DOWNLOAD   = "download"
	TAG_STREAM     = "getdatapoints"
	TAG_GETRECORDS = "getrecords"
)

var (
	ErrBlocked              = status.Errorf(codes.Unavailable, "blocked")
	ErrStreamNotAvailable   = status.Errorf(codes.NotFound, "stream not available")
	ErrPropertyNotAvailable = status.Errorf(codes.NotFound, "property not available")
	ErrCaptureNotAvailable  = status.Errorf(codes.NotFound, "capture not available")
	ErrEventNotAvailable    = status.Errorf(codes.NotFound, "event not available")
)

func (s *Service) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	methodAttr := metrics.KeyGrpcMethod.String("GetSession")
	s.smetrics.GrpcReqCount.Add(ctx, 1, metric.WithAttributes(methodAttr))
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributes(methodAttr))
	}()

	return &pb.GetSessionResponse{
		Uuid: s.session.String(),
	}, nil
}

func (s *Service) GetProperties(req *pb.GetPropertiesRequest, srv pb.Telemetry_GetPropertiesServer) error {
	methodAttr := metrics.KeyGrpcMethod.String("GetProperties")
	s.smetrics.GrpcReqCount.Add(srv.Context(), 1, metric.WithAttributes(methodAttr))
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(srv.Context(), time.Since(startTime).Milliseconds(), metric.WithAttributes(methodAttr))
	}()

	properties := s.properties.copyProperties()
	for _, v := range properties {
		if err := srv.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) exportStreamToGrpcServer(stream *stream.Stream, _since uint32, srv grpc.ServerStreamingServer[pb.StreamSegment]) (int, error) {
	segmentCount := 0
	since := int(_since)
	for {
		segments := stream.Segments(since, 128)
		if len(segments) == 0 {
			break
		}
		segmentCount += len(segments)
		since = segments[len(segments)-1].SeqN + 1
		for _, segment := range segments {
			err := srv.Send(&pb.StreamSegment{
				SequenceNumber: uint32(segment.SeqN),
				Data:           segment.Data,
			})
			if err != nil {
				return segmentCount, err
			}
		}
	}

	return segmentCount, nil
}

func (s *Service) GetMetrics(req *pb.GetMetricsRequest, srv grpc.ServerStreamingServer[pb.StreamSegment]) error {
	segmentCount, err := s.exportStreamToGrpcServer(s.metrics.stream, req.GetSequenceNumberSince(), srv)
	methodAttr := metrics.KeyGrpcMethod.String("GetStream")
	s.smetrics.GrpcStreamSegRet.Record(srv.Context(), int64(segmentCount), metric.WithAttributes(methodAttr))
	return err
}

func (s *Service) GetEventDescriptors(req *pb.GetEventDescriptorsRequest, srv grpc.ServerStreamingServer[pb.EventDescriptor]) error {
	descriptors := s.events.getEventDescriptors()
	for _, descriptor := range descriptors {
		if err := srv.Send(descriptor); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetEvents(req *pb.GetEventsRequest, srv grpc.ServerStreamingServer[pb.StreamSegment]) error {
	stream := s.events.getEventStreamById(eventId(req.GetEventId()))
	if stream == nil {
		return ErrEventNotAvailable
	}

	segmentCount, err := s.exportStreamToGrpcServer(s.metrics.stream, req.GetSequenceNumberSince(), srv)
	methodAttr := metrics.KeyGrpcMethod.String("GetEvents")
	s.smetrics.GrpcStreamSegRet.Record(srv.Context(), int64(segmentCount), metric.WithAttributes(methodAttr))
	return err
}
