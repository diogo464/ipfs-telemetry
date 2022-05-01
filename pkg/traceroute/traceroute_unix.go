//go:build linux || darwin
// +build linux darwin

package traceroute

import "os/exec"

var providerTraceroute = provider{
	name: "traceroute",
	fn: func(destination string) ([]byte, error) {
		return exec.Command("traceroute", destination).Output()
	},
}

var providerTracepath4 = provider{
	name: "tracepath4",
	fn: func(destination string) ([]byte, error) {
		return exec.Command("tracepath", "-n", "-4", destination).Output()
	},
}

var providerTracepath6 = provider{
	name: "tracepath6",
	fn: func(destination string) ([]byte, error) {
		return exec.Command("tracepath", "-n", "-6", destination).Output()
	},
}

func platformTrace(destination string) (*Result, error) {
	return runAndCollectResults(destination,
		providerTraceroute,
		providerTracepath4,
		providerTracepath6,
	)
}
