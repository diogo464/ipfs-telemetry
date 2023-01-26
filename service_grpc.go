package telemetry

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/metrics"
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
	s.smetrics.GrpcReqCount.Add(ctx, 1, methodAttr)
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(ctx, time.Since(startTime).Milliseconds(), methodAttr)
	}()

	return &pb.GetSessionResponse{
		Uuid: s.session.String(),
	}, nil
}

func (s *Service) GetProperties(req *pb.GetPropertiesRequest, srv pb.Telemetry_GetPropertiesServer) error {
	methodAttr := metrics.KeyGrpcMethod.String("GetProperties")
	s.smetrics.GrpcReqCount.Add(srv.Context(), 1, methodAttr)
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(srv.Context(), time.Since(startTime).Milliseconds(), methodAttr)
	}()

	properties := s.properties.copyProperties()
	for _, v := range properties {
		if err := srv.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetStreamDescriptors(ctx context.Context, req *pb.GetStreamDescriptorsRequest) (*pb.GetStreamDescriptorsResponse, error) {
	methodAttr := metrics.KeyGrpcMethod.String("GetStreamDescriptors")
	s.smetrics.GrpcReqCount.Add(ctx, 1, methodAttr)
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(ctx, time.Since(startTime).Milliseconds(), methodAttr)
	}()

	descriptors := s.streams.copyDescriptors()
	return &pb.GetStreamDescriptorsResponse{
		StreamDescriptors: descriptors,
	}, nil
}

func (s *Service) GetStream(req *pb.GetStreamRequest, srv pb.Telemetry_GetStreamServer) error {
	methodAttr := metrics.KeyGrpcMethod.String("GetStream")
	s.smetrics.GrpcReqCount.Add(srv.Context(), 1, methodAttr)
	startTime := time.Now()
	defer func() {
		s.smetrics.GrpcReqDur.Record(srv.Context(), time.Since(startTime).Milliseconds(), methodAttr)
	}()

	streamId := StreamId(req.GetStreamId())
	sstream := s.streams.get(streamId)
	if sstream == nil {
		return ErrStreamNotAvailable
	}
	stream := sstream.stream

	segmentCount := 0
	since := int(req.GetSequenceNumberSince())
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
				return err
			}
		}
	}

	s.smetrics.GrpcStreamSegRet.Record(srv.Context(), int64(segmentCount), methodAttr)
	return nil
}
