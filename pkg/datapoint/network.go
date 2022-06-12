package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

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

func NetworkSerialize(in *Network, stream *telemetry.Stream) error {
	byprotocol := make(map[string]*pb.Network_Stats)
	for k, v := range in.StatsByProtocol {
		byprotocol[string(k)] = pbutils.MetricsStatsToPB(&v)
	}
	bypeer := make(map[string]*pb.Network_Stats)
	for k, v := range in.StatsByPeer {
		bypeer[k.Pretty()] = pbutils.MetricsStatsToPB(&v)
	}

	inpb := &pb.Network{
		Timestamp:       pbutils.TimeToPB(&in.Timestamp),
		Addresses:       pbutils.MultiAddrsToPB(in.Addresses),
		StatsOverall:    pbutils.MetricsStatsToPB(&in.Overall),
		StatsByProtocol: byprotocol,
		StatsByPeer:     bypeer,
		NumConns:        in.NumConns,
		LowWater:        in.LowWater,
		HighWater:       in.HighWater,
	}

	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func NetworkDeserialize(in []byte) (*Network, error) {
	var inpb pb.Network
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	byprotocol := make(map[protocol.ID]metrics.Stats, len(inpb.GetStatsByProtocol()))
	for k, v := range inpb.GetStatsByProtocol() {
		byprotocol[protocol.ID(k)] = pbutils.MetricsStatsFromPB(v)
	}
	bypeer := make(map[peer.ID]metrics.Stats, len(inpb.GetStatsByPeer()))
	for k, v := range inpb.GetStatsByPeer() {
		p, err := peer.Decode(k)
		if err != nil {
			return nil, err
		}
		bypeer[p] = pbutils.MetricsStatsFromPB(v)
	}
	addresses := make([]multiaddr.Multiaddr, 0, len(inpb.GetAddresses()))
	for _, a := range inpb.GetAddresses() {
		addr, err := multiaddr.NewMultiaddr(a)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}

	return &Network{
		Timestamp:       pbutils.TimeFromPB(inpb.GetTimestamp()),
		Addresses:       addresses,
		Overall:         pbutils.MetricsStatsFromPB(inpb.GetStatsOverall()),
		StatsByProtocol: byprotocol,
		StatsByPeer:     bypeer,
		NumConns:        inpb.GetNumConns(),
		LowWater:        inpb.GetLowWater(),
		HighWater:       inpb.GetHighWater(),
	}, nil
}
