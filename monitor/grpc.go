package monitor

// import (
// 	"context"

// 	"github.com/diogo464/telemetry/internal/actionqueue"
// 	"github.com/diogo464/telemetry/monitor/pb"
// 	"github.com/gogo/protobuf/types"
// 	"github.com/libp2p/go-libp2p-core/peer"
// )

// func (s *Monitor) Discover(ctx context.Context, req *pb.DiscoverRequest) (*types.Empty, error) {
// 	p, err := peer.Decode(req.Peer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.caction <- actionqueue.Now(&action{
// 		kind: ActionDiscover,
// 		pid:  p,
// 	})
// 	return &types.Empty{}, nil
// }
