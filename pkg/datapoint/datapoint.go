package datapoint

import (
	"time"

	"github.com/diogo464/telemetry"
)

type Datapoint interface{}

var Decoders = map[string]telemetry.StreamDecoder[Datapoint]{
	BitswapName:         datapointDecoder(BitswapDeserialize),
	ConnectionsName:     datapointDecoder(ConnectionsDeserialize),
	KademliaName:        datapointDecoder(KademliaDeserialize),
	KademliaHandlerName: datapointDecoder(KademliaHandlerDeserialize),
	KademliaQueryName:   datapointDecoder(KademliaQueryDeserialize),
	NetworkName:         datapointDecoder(NetworkDeserialize),
	PingName:            datapointDecoder(PingDeserialize),
	ResourceName:        datapointDecoder(ResourcesDeserialize),
	RoutingTableName:    datapointDecoder(RoutingTableDeserialize),
	StorageName:         datapointDecoder(StorageDeserialize),
	TraceRouteName:      datapointDecoder(TraceRouteDeserialize),
}

func NewTimestamp() time.Time {
	return time.Now().UTC()
}

func datapointDecoder[T any](decoder telemetry.StreamDecoder[T]) telemetry.StreamDecoder[Datapoint] {
	return func(d []byte) (Datapoint, error) {
		return decoder(d)
	}
}
