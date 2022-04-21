package config

import "time"

type Config struct {
	Ping             Ping
	NetworkCollector NetworkCollector
	RoutingTable     RoutingTable
	Resources        Resources
	Bitswap          Bitswap
	Storage          Storage
	Kademlia         Kademlia
	TraceRoute       TraceRoute
	Window           Window
}

type Ping struct {
	Interval int
	Timeout  int
	Count    int
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
	Interval int
}

func SecondsToDuration(s int, def time.Duration) time.Duration {
	if s == 0 {
		return def
	}
	return time.Duration(int64(s)) * time.Second
}
