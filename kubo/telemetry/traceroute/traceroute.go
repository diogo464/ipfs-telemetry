package traceroute

import (
	"context"
	"fmt"
)

var ErrNoProviderAvailable = fmt.Errorf("no traceroute provider available")

type Result struct {
	Provider string
	Output   []byte
}

func Trace(ctx context.Context, destination string) (*Result, error) {
	return platformTrace(ctx, destination)
}

type provider struct {
	name string
	fn   func(ctx context.Context, destination string) ([]byte, error)
}

func runAndCollectResults(ctx context.Context, destination string, providers ...provider) (*Result, error) {
	var err error = ErrNoProviderAvailable
	for _, p := range providers {
		if output, err := p.fn(ctx, destination); err == nil {
			return &Result{
				Provider: p.name,
				Output:   output,
			}, nil
		}
	}
	return nil, err
}
