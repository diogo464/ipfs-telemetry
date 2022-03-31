package telemetry

import (
	"fmt"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

var ERR_INVALIDS_SNAPSHOT_TYPE = fmt.Errorf("invalid snapshot type")

type Snapshot interface {
}

type PingSnapshot struct {
	Timestamp   time.Time       `json:"time"`
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func PingFromPB(in *pb.Snapshot_Ping) (*PingSnapshot, error) {
	source, err := addrInfoFromPB(in.Source)
	if err != nil {
		return nil, err
	}
	dest, err := addrInfoFromPB(in.Destination)
	if err != nil {
		return nil, err
	}
	durations := make([]time.Duration, 0, len(in.Durations))
	for _, dur := range in.Durations {
		durations = append(durations, dur.AsDuration())
	}
	return &PingSnapshot{
		Timestamp:   in.GetTimestamp().AsTime(),
		Source:      source,
		Destination: dest,
		Durations:   durations,
	}, nil
}

type RoutingTableSnapshot struct {
	Timestamp time.Time `json:"time"`
	Buckets   []uint32  `json:"buckets"`
}

func RoutingTableFromPB(in *pb.Snapshot_RoutingTable) (*RoutingTableSnapshot, error) {
	return &RoutingTableSnapshot{
		Timestamp: in.GetTimestamp().AsTime(),
		Buckets:   []uint32{},
	}, nil
}

type NetworkSnapshot struct {
	Timestamp time.Time `json:"time"`
	TotalIn   uint64    `json:"total_in"`
	TotalOut  uint64    `json:"total_out"`
	RateIn    uint64    `json:"rate_in"`
	RateOut   uint64    `json:"rate_out"`
	NumConns  uint32    `json:"num_conns"`
	LowWater  uint32    `json:"low_water"`
	HighWater uint32    `json:"high_water"`
}

func NetworkFromPB(in *pb.Snapshot_Network) (*NetworkSnapshot, error) {
	return &NetworkSnapshot{
		Timestamp: in.GetTimestamp().AsTime(),
		TotalIn:   in.GetTotalIn(),
		TotalOut:  in.GetTotalOut(),
		RateIn:    in.GetRateIn(),
		RateOut:   in.GetRateOut(),
		NumConns:  in.GetNumConns(),
		LowWater:  in.GetLowWater(),
		HighWater: in.GetHighWater(),
	}, nil
}

func addrInfoFromPB(in *pb.AddrInfo) (peer.AddrInfo, error) {
	i, err := peer.Decode(in.Id)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	a := make([]multiaddr.Multiaddr, 0, len(in.Addrs))
	for _, addr := range in.Addrs {
		x, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return peer.AddrInfo{}, err
		}
		a = append(a, x)
	}

	return peer.AddrInfo{ID: i, Addrs: a}, nil
}
