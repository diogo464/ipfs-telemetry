package bitswap

import "sync"

type BitswapDiscoveryStats struct {
	Succeeded uint
	Failed    uint
}

type BitswapTelemetry struct {
	lock      sync.Mutex
	discovery BitswapDiscoveryStats
}

func NewBitswapTelemetry() *BitswapTelemetry {
	return &BitswapTelemetry{}
}

func (bt *BitswapTelemetry) GetDiscoveryStats() BitswapDiscoveryStats {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	return bt.discovery
}

func (bt *BitswapTelemetry) AddDiscoverySuccess() {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	bt.discovery.Succeeded += 1
}

func (bt *BitswapTelemetry) AddDiscoveryFailure() {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	bt.discovery.Failed += 1
}
