//go:build windows
// +build windows

package traceroute

import (
	"context"
	"os/exec"
)

var providerTracert = provider{
	name: "tracert",
	fn: func(ctx context.Context, destination string) ([]byte, error) {
		return exec.CommandContext(ctx, "tracert", destination).Output()
	},
}

func platformTrace(ctx context.Context, destination string) (*Result, error) {
	return runAndCollectResults(ctx, destination, providerTracert)
}
