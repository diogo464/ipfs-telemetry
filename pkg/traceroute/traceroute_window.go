//go:build windows
// +build windows

package traceroute

import "os/exec"

var providerTracert = provider{
	name: "tracert",
	fn: func(destination string) ([]byte, error) {
		return exec.Command("tracert", destination).Output()
	},
}

func platformTrace(destination string) (*Result, error) {
	return runAndCollectResults(destination, providerTracert)
}
