package traceroute

import (
	"fmt"
)

var ErrNoProviderAvailable = fmt.Errorf("no traceroute provider available")

type Result struct {
	Provider string
	Output   []byte
}

// TODO: add context
func Trace(destination string) (*Result, error) {
	return platformTrace(destination)
}

type provider struct {
	name string
	fn   func(destination string) ([]byte, error)
}

func runAndCollectResult(destination string, providers ...provider) (*Result, error) {
	var err error = ErrNoProviderAvailable
	for _, p := range providers {
		if output, err := p.fn(destination); err == nil {
			return &Result{
				Provider: p.name,
				Output:   output,
			}, nil
		}
	}
	return nil, err
}
