package telemetry

import (
	"context"
	"fmt"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
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
	return &pb.GetSessionResponse{
		Uuid: s.session.String(),
	}, nil
}

func (s *Service) GetPropertyDescriptors(req *pb.GetPropertyDescriptorsRequest, srv pb.Telemetry_GetPropertyDescriptorsServer) error {
	descriptors := s.properties.copyDescriptors()
	for _, desc := range descriptors {
		if err := srv.Send(desc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetProperties(req *pb.GetPropertiesRequest, srv pb.Telemetry_GetPropertiesServer) error {
	properties := s.properties.copyProperties()
	for _, v := range properties {
		if err := srv.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetMetricDescriptors(req *pb.GetMetricDescriptorsRequest, srv pb.Telemetry_GetMetricDescriptorsServer) error {
	descriptors := s.metrics.copyDescriptors()
	for _, desc := range descriptors {
		if err := srv.Send(desc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetCaptureDescriptors(req *pb.GetCaptureDescriptorsRequest, srv pb.Telemetry_GetCaptureDescriptorsServer) error {
	descriptors := s.captures.copyDescriptors()
	for _, v := range descriptors {
		if err := srv.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetEventDescriptors(req *pb.GetEventDescriptorsRequest, srv pb.Telemetry_GetEventDescriptorsServer) error {
	descriptors := s.events.copyDescriptors()
	for _, desc := range descriptors {
		if err := srv.Send(desc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetStream(req *pb.GetStreamRequest, srv pb.Telemetry_GetStreamServer) error {
	streamId := StreamId(req.GetStreamId())
	stream := s.streams.get(streamId)
	if stream == nil {
		return ErrStreamNotAvailable
	}
	return grpcSendStreamSegments(stream.stream, int(req.GetSequenceNumberSince()), srv)
}

func grpcSendStreamSegments(stream *stream.Stream, since int, srv grpcSegmentSender) error {
	fmt.Println("Sending segments since", since)
	for {
		segments := stream.Segments(since, 128)
		if len(segments) == 0 {
			break
		}
		since = segments[len(segments)-1].SeqN + 1
		fmt.Println("Sending", len(segments), "segments")
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
	return nil
}
