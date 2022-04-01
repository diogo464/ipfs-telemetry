package bitswap

import (
	"go.uber.org/atomic"
)

type SessionTelemetryState struct {
	// we have already collected the success/failure of the discovery
	complete    *atomic.Bool
	bstelemetry *BitswapTelemetry
}

func NewSessionTelemetryState(bstelemetry *BitswapTelemetry) *SessionTelemetryState {
	return &SessionTelemetryState{
		complete:    atomic.NewBool(false),
		bstelemetry: bstelemetry,
	}
}

func (st *SessionTelemetryState) DiscoverySucceeded() {
	if st.complete.CAS(false, true) {
		st.bstelemetry.AddDiscoverySuccess()
	}
}

func (st *SessionTelemetryState) DiscoveryFailed() {
	if st.complete.CAS(false, true) {
		st.bstelemetry.AddDiscoveryFailure()
	}
}
