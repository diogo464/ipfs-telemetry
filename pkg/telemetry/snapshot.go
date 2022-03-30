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

type SnapshotHeader struct {
	time.Time `json:"time"`
}

type PingSnapshot struct {
	Header      SnapshotHeader  `json:"header"`
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

type RoutingTableSnapshot struct {
	Header  SnapshotHeader `json:"header"`
	Buckets []uint32       `json:"buckets"`
}

type NetworkSnapshot struct {
	Header    SnapshotHeader `json:"header"`
	TotalIn   uint64         `json:"total_in"`
	TotalOut  uint64         `json:"total_out"`
	RateIn    uint64         `json:"rate_in"`
	RateOut   uint64         `json:"rate_out"`
	NumConns  uint32         `json:"num_conns"`
	LowWater  uint32         `json:"low_water"`
	HighWater uint32         `json:"high_water"`
}

func snapshotFromPB(snapshot *pb.Snapshot) (Snapshot, error) {
	// TODO: snapshot data

	switch snapshot.GetBody().(type) {
	case *pb.Snapshot_Ping_:
		s := snapshot.GetPing()
		source, err := parsePeerAddrInfo(s.SourcePid, s.SourceAddrs)
		if err != nil {
			return nil, err
		}
		destination, err := parsePeerAddrInfo(s.DestinationPid, s.DestinationAddrs)
		if err != nil {
			return nil, err
		}
		durations := make([]time.Duration, 0, len(s.Durations))
		for _, d := range s.Durations {
			durations = append(durations, d.AsDuration())
		}
		return &PingSnapshot{
			// SnapshotData: SnapshotData{},
			Source:      source,
			Destination: destination,
			Durations:   durations,
		}, nil
	case *pb.Snapshot_RoutingTable_:
		s := snapshot.GetRoutingTable()
		return &RoutingTableSnapshot{
			//SnapshotData: SnapshotData{},
			Buckets: s.Buckets,
		}, nil
	case *pb.Snapshot_Network_:
		s := snapshot.GetNetwork()
		return &NetworkSnapshot{
			// SnapshotData: SnapshotData{},
			TotalIn:   s.TotalIn,
			TotalOut:  s.TotalOut,
			RateIn:    s.RateIn,
			RateOut:   s.RateOut,
			NumConns:  s.NumConns,
			LowWater:  s.LowWater,
			HighWater: s.HighWater,
		}, nil
	default:
		return nil, ERR_INVALIDS_SNAPSHOT_TYPE
	}
}

func parsePeerAddrInfo(pid string, addrs []string) (peer.AddrInfo, error) {
	i, err := peer.Decode(pid)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	a := make([]multiaddr.Multiaddr, 0, len(addrs))
	for _, addr := range addrs {
		x, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return peer.AddrInfo{}, err
		}
		a = append(a, x)
	}

	return peer.AddrInfo{ID: i, Addrs: a}, nil
}
