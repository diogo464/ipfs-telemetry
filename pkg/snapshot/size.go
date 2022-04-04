package snapshot

// Size estimates for snapshots
// TODO: remove later

const (
	peerIdSize       = 64
	multiAddrSize    = 48
	peerAddrInfoSize = peerIdSize + multiAddrSize*8
	durationSize     = 8
	timestampSize    = 24
	metricsStatsSize = 8 + 8 + 8 + 8
)

func SnapshotSize(s Snapshot) int {
	switch v := s.(type) {
	case *Ping:
		return PingSize(v)
	case *RoutingTable:
		return RoutingTableSize(v)
	case *Network:
		return NetworkSize(v)
	case *Resources:
		return ResourcesSize(v)
	case *TraceRoute:
		return TraceRouteSize(v)
	case *Kademlia:
		return KademliaSize(v)
	case *KademliaQuery:
		return KademliaQuerySize(v)
	case *Bitswap:
		return BitswapSize(v)
	case *Storage:
		return StorageSize(v)
	default:
		panic("unimplemented")
	}
}

func PingSize(s *Ping) int {
	return timestampSize + peerAddrInfoSize*2 + len(s.Durations)*durationSize
}

func RoutingTableSize(s *RoutingTable) int {
	totalPeers := 0
	for _, b := range s.Buckets {
		totalPeers += len(b)
	}
	return timestampSize + totalPeers*peerIdSize
}

func NetworkSize(s *Network) int {
	return timestampSize + metricsStatsSize + len(s.StatsByProtocol)*metricsStatsSize + len(s.StatsByPeer)*(metricsStatsSize+peerIdSize) + 4*3
}

func ResourcesSize(s *Resources) int {
	return timestampSize + 4*2 + 8*3
}

func TraceRouteSize(s *TraceRoute) int {
	return timestampSize + 2*peerAddrInfoSize + len(s.Provider) + len(s.Output)
}

func KademliaSize(s *Kademlia) int {
	return timestampSize + len(s.MessagesIn)*8 + len(s.MessagesOut)*8
}

func KademliaQuerySize(s *KademliaQuery) int {
	return timestampSize + peerIdSize + 4 + durationSize
}

func BitswapSize(s *Bitswap) int {
	return timestampSize + 4*4
}

// func IpnsSize(s*Ipns) int {}

func StorageSize(s *Storage) int {
	return timestampSize + 8*3
}
