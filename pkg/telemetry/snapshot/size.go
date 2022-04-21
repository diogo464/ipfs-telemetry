package snapshot

// Size estimates for snapshots
// TODO: remove later

//const (
//	peerIdSize       = 64
//	multiAddrSize    = 48
//	peerAddrInfoSize = peerIdSize + multiAddrSize*8
//	durationSize     = 8
//	timestampSize    = 24
//	metricsStatsSize = 8 + 8 + 8 + 8
//)
//
//func SnapshotSize(s Snapshot) uint32 {
//	switch v := s.(type) {
//	case *Ping:
//		return PingSize(v)
//	case *RoutingTable:
//		return RoutingTableSize(v)
//	case *Network:
//		return NetworkSize(v)
//	case *Resources:
//		return ResourcesSize(v)
//	case *TraceRoute:
//		return TraceRouteSize(v)
//	case *Kademlia:
//		return KademliaSize(v)
//	case *KademliaQuery:
//		return KademliaQuerySize(v)
//	case *KademliaHandler:
//		return KademliaHandlerSize(v)
//	case *Bitswap:
//		return BitswapSize(v)
//	case *Storage:
//		return StorageSize(v)
//	case *Window:
//		return WindowSize(v)
//	default:
//		panic("unimplemented")
//	}
//}
//
//func PingSize(s *Ping) uint32 {
//	return timestampSize + peerAddrInfoSize*2 + uint32(len(s.Durations))*durationSize
//}
//
//func RoutingTableSize(s *RoutingTable) uint32 {
//	var totalPeers uint32 = 0
//	for _, b := range s.Buckets {
//		totalPeers += uint32(len(b))
//	}
//	return timestampSize + totalPeers*peerIdSize
//}
//
//func NetworkSize(s *Network) uint32 {
//	return timestampSize + metricsStatsSize + uint32(len(s.StatsByProtocol))*metricsStatsSize + /*len(s.StatsByPeer)*(metricsStatsSize+peerIdSize) */ +4*3
//}
//
//func ResourcesSize(s *Resources) uint32 {
//	return timestampSize + 4*2 + 8*3
//}
//
//func TraceRouteSize(s *TraceRoute) uint32 {
//	return timestampSize + 2*peerAddrInfoSize + uint32(len(s.Provider)) + uint32(len(s.Output))
//}
//
//func KademliaSize(s *Kademlia) uint32 {
//	return timestampSize + uint32(len(s.MessagesIn))*8 + uint32(len(s.MessagesOut))*8
//}
//
//func KademliaQuerySize(s *KademliaQuery) uint32 {
//	return timestampSize + peerIdSize + 4 + durationSize
//}
//
//func KademliaHandlerSize(s *KademliaHandler) uint32 {
//	return timestampSize + 4 + 2*durationSize
//}
//
//func BitswapSize(s *Bitswap) uint32 {
//	return timestampSize + 4*4
//}
//
//// func IpnsSize(s*Ipns) uint32 {}
//
//func StorageSize(s *Storage) uint32 {
//	return timestampSize + 8*3
//}
//
//func WindowSize(s *Window) uint32 {
//	// 18 -> 8 bytes of uuint3264 + ~10 bytes of name
//	return timestampSize + uint32(len(s.SnapshotCount))*18 + uint32(len(s.SnapshotMemory))*18
//}
