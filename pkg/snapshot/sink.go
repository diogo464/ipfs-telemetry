package snapshot

type Sink interface {
	PushPing(*Ping)
	PushRoutingTable(*RoutingTable)
	PushNetwork(*Network)
	PushResources(*Resources)
	PushBitswap(*Bitswap)
	PushStorage(*Storage)
}
