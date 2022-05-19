package config

import "time"

type Config struct {
	Ping             Ping
	Connections      Connections
	NetworkCollector NetworkCollector
	RoutingTable     RoutingTable
	Resources        Resources
	Bitswap          Bitswap
	Storage          Storage
	Kademlia         Kademlia
	TraceRoute       TraceRoute
	Window           Window
	Relay            Relay
}

func Default() Config {
	return Config{
		Ping: Ping{
			Interval: 5,
			Timeout:  10,
			Count:    5,
		},
		Connections: Connections{
			Interval: 60,
		},
		NetworkCollector: NetworkCollector{
			Interval:                30,
			BandwidthByPeerInterval: 5 * 60,
		},
		RoutingTable: RoutingTable{
			Interval: 60,
		},
		Resources: Resources{
			Interval: 10,
		},
		Bitswap: Bitswap{
			Interval: 30,
		},
		Storage: Storage{
			Interval: 60,
		},
		Kademlia: Kademlia{
			Interval: 30,
		},
		TraceRoute: TraceRoute{
			Interval: 5,
		},
		Window: Window{
			Interval:   5,
			Duration:   30 * 60,
			EventCount: 128 * 1024,
		},
		Relay: Relay{
			Interval: 30,
		},
	}
}

type Ping struct {
	Interval int
	Timeout  int
	Count    int
}

type Connections struct {
	Interval int
}

type NetworkCollector struct {
	Interval                int
	BandwidthByPeerInterval int
}

type RoutingTable struct {
	Interval int
}

type Resources struct {
	Interval int
}

type Bitswap struct {
	Interval int
}

type Storage struct {
	Interval int
}

type Kademlia struct {
	Interval int
}

type TraceRoute struct {
	Interval int
}

type Window struct {
	Interval   int
	Duration   int
	EventCount int
}

type Relay struct {
	Interval int
}

func SecondsToDuration(s int, def int) time.Duration {
	if s == 0 {
		return time.Duration(int64(def)) * time.Second
	}
	return time.Duration(int64(s)) * time.Second
}
