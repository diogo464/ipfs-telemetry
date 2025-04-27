package main

import (
	"encoding/json"
	"os"

	"github.com/diogo464/telemetry/walker"
)

var _ (walker.Observer) = (*fileObserver)(nil)

type fileObserver struct {
	success *os.File
	failure *os.File
}

func newFileObserver(successPath string, failurePath string) (*fileObserver, error) {
	success, err := os.Create(successPath)
	if err != nil {
		return nil, err
	}
	failure, err := os.Create(failurePath)
	if err != nil {
		return nil, err
	}
	return &fileObserver{success, failure}, nil
}

// ObserveError implements walker.Observer.
func (f *fileObserver) ObserveError(e *walker.Error) {
	if m, err := json.Marshal(e); err == nil {
		f.failure.Write(m)
		f.failure.Write([]byte("\n"))
		f.failure.Sync()
	}
}

// ObservePeer implements walker.Observer.
func (f *fileObserver) ObservePeer(p *walker.Peer) {
	if m, err := json.Marshal(p); err == nil {
		f.success.Write(m)
		f.success.Write([]byte("\n"))
		f.success.Sync()
	}
}
