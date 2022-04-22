package datapoint

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/datapoint"
	"git.d464.sh/adc/telemetry/pkg/telemetry/pbutils"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*Network)(nil)

const NetworkName = "network"

type Network struct {
	Timestamp       time.Time                     `json:"timestamp"`
	Addresses       []multiaddr.Multiaddr         `json:"addresses"`
	Overall         metrics.Stats                 `json:"overall"`
	StatsByProtocol map[protocol.ID]metrics.Stats `json:"stats_by_protocol"`
	StatsByPeer     map[peer.ID]metrics.Stats     `json:"stats_by_peer"`
	NumConns        uint32                        `json:"numconns"`
	LowWater        uint32                        `json:"lowwater"`
	HighWater       uint32                        `json:"highwater"`
}

func (*Network) sealed()                   {}
func (*Network) GetName() string           { return NetworkName }
func (n *Network) GetTimestamp() time.Time { return n.Timestamp }
func (n *Network) GetSizeEstimate() uint32 {
	return estimateTimestampSize + uint32(len(n.Addresses))*estimateMultiAddrSize +
		estimateMetricsStatsSize +
		uint32(len(n.StatsByProtocol)) + (estimateProtocolIdSize + estimateMetricsStatsSize) +
		uint32(len(n.StatsByPeer)) + (estimatePeerIdSize + estimateMetricsStatsSize) +
		3*4
}
func (n *Network) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Network{
			Network: NetworkToPB(n),
		},
	}
}

func NetworkFromPB(in *pb.Network) (*Network, error) {
	byprotocol := make(map[protocol.ID]metrics.Stats, len(in.GetStatsByProtocol()))
	for k, v := range in.GetStatsByProtocol() {
		byprotocol[protocol.ID(k)] = pbutils.MetricsStatsFromPB(v)
	}
	bypeer := make(map[peer.ID]metrics.Stats, len(in.GetStatsByPeer()))
	for k, v := range in.GetStatsByPeer() {
		p, err := peer.Decode(k)
		if err != nil {
			return nil, err
		}
		bypeer[p] = pbutils.MetricsStatsFromPB(v)
	}
	addresses := make([]multiaddr.Multiaddr, 0, len(in.GetAddresses()))
	for _, a := range in.GetAddresses() {
		addr, err := multiaddr.NewMultiaddr(a)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}

	return &Network{
		Timestamp:       in.GetTimestamp().AsTime(),
		Addresses:       addresses,
		Overall:         pbutils.MetricsStatsFromPB(in.GetStatsOverall()),
		StatsByProtocol: byprotocol,
		StatsByPeer:     bypeer,
		NumConns:        in.GetNumConns(),
		LowWater:        in.GetLowWater(),
		HighWater:       in.GetHighWater(),
	}, nil
}

func NetworkToPB(n *Network) *pb.Network {
	byprotocol := make(map[string]*pb.Network_Stats)
	for k, v := range n.StatsByProtocol {
		byprotocol[string(k)] = pbutils.MetricsStatsToPB(&v)
	}
	bypeer := make(map[string]*pb.Network_Stats)
	for k, v := range n.StatsByPeer {
		bypeer[k.Pretty()] = pbutils.MetricsStatsToPB(&v)
	}

	return &pb.Network{
		Timestamp:       timestamppb.New(n.Timestamp),
		Addresses:       pbutils.MultiAddrsToPB(n.Addresses),
		StatsOverall:    pbutils.MetricsStatsToPB(&n.Overall),
		StatsByProtocol: byprotocol,
		StatsByPeer:     bypeer,
		NumConns:        n.NumConns,
		LowWater:        n.LowWater,
		HighWater:       n.HighWater,
	}
}

func NetworkArrayToPB(in []*Network) []*pb.Network {
	out := make([]*pb.Network, 0, len(in))
	for _, p := range in {
		out = append(out, NetworkToPB(p))
	}
	return out
}
