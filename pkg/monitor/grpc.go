package monitor

import (
	"context"

	"github.com/diogo464/telemetry/pkg/actionqueue"
	pb "github.com/diogo464/telemetry/pkg/proto/monitor"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Monitor) Discover(ctx context.Context, req *pb.DiscoverRequest) (*emptypb.Empty, error) {
	p, err := peer.Decode(req.Peer)
	if err != nil {
		return nil, err
	}
	s.caction <- actionqueue.Now(&action{
		kind: ActionDiscover,
		pid:  p,
	})
	return &emptypb.Empty{}, nil
}
