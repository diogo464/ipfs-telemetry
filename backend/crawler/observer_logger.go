package crawler

import (
	"github.com/diogo464/telemetry/walker"
	"go.uber.org/zap"
)

var _ (walker.Observer) = (*loggerObserver)(nil)

type loggerObserver struct {
	l *zap.Logger
}

func newLoggerObserver(l *zap.Logger) *loggerObserver {
	return &loggerObserver{l}
}

// ObserveError implements walker.Observer
func (o *loggerObserver) ObserveError(e *walker.Error) {
	o.l.Warn("error", zap.String("peer", e.ID.String()), zap.Error(e.Err))
}

// ObservePeer implements walker.Observer
func (o *loggerObserver) ObservePeer(c *walker.Peer) {
	o.l.Info("observing peer", zap.String("peer", c.ID.String()))
}
