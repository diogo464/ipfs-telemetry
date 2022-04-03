package measurements

import "go.uber.org/atomic"

var bitswap Bitswap = nil

type Bitswap interface {
	IncDiscoverySuccess()
	IncDiscoveryFailure()
}

func BitswapRegister(b Bitswap) {
	if bitswap != nil {
		panic("should not happen")
	}
	bitswap = b
}

func WithBitswap(fn func(Bitswap)) {
	if bitswap != nil {
		fn(bitswap)
	}
}

type SessionTelemetryState struct {
	// we have already collected the success/failure of the discovery
	complete *atomic.Bool
}

func NewSessionTelemetryState() *SessionTelemetryState {
	return &SessionTelemetryState{
		complete: atomic.NewBool(false),
	}
}

func (st *SessionTelemetryState) DiscoverySucceeded() {
	if st.complete.CAS(false, true) {
		WithBitswap(func(b Bitswap) { b.IncDiscoverySuccess() })
	}
}

func (st *SessionTelemetryState) DiscoveryFailed() {
	if st.complete.CAS(false, true) {
		WithBitswap(func(b Bitswap) { b.IncDiscoveryFailure() })
	}
}
