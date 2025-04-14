//go:build linux || darwin || freebsd || openbsd
// +build linux darwin freebsd openbsd

package traceroute

import (
	"context"
	"os/exec"
)

var providerTraceroute = provider{
	name: "traceroute",
	fn: func(ctx context.Context, destination string) ([]byte, error) {
		return exec.CommandContext(ctx, "traceroute", destination).Output()
	},
}

var providerTracepath4 = provider{
	name: "tracepath4",
	fn: func(ctx context.Context, destination string) ([]byte, error) {
		return exec.CommandContext(ctx, "tracepath", "-n", "-4", destination).Output()
	},
}

var providerTracepath6 = provider{
	name: "tracepath6",
	fn: func(ctx context.Context, destination string) ([]byte, error) {
		return exec.CommandContext(ctx, "tracepath", "-n", "-6", destination).Output()
	},
}

func platformTrace(ctx context.Context, destination string) (*Result, error) {
	return runAndCollectResults(ctx, destination,
		providerTraceroute,
		providerTracepath4,
		providerTracepath6,
	)
}
