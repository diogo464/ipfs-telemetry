package crawler

import "github.com/diogo464/telemetry/walker"

type Observer interface {
	walker.Observer
	CrawlBegin()
	CrawlEnd()
}

var _ (Observer) = (*walkerObserverBridge)(nil)

type walkerObserverBridge struct {
	observer walker.Observer
}

func (o *walkerObserverBridge) ObservePeer(p *walker.Peer) {
	o.observer.ObservePeer(p)
}
func (o *walkerObserverBridge) ObserveError(e *walker.Error) {
	o.observer.ObserveError(e)
}
func (o *walkerObserverBridge) CrawlBegin() {}
func (o *walkerObserverBridge) CrawlEnd()   {}
func newWalkerObserverBridge(observer walker.Observer) *walkerObserverBridge {
	return &walkerObserverBridge{observer}
}
