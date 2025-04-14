package telemetry

import (
	"fmt"
	"net"

	"github.com/multiformats/go-multiaddr"
)

func getFirstPublicAddressFromMultiaddrs(in []multiaddr.Multiaddr) (net.IP, error) {
	public := getPublicAddressesFromMultiaddrs(in)
	if len(public) == 0 {
		return nil, fmt.Errorf("no public address found")
	} else {
		return public[0], nil
	}
}

func getPublicAddressesFromMultiaddrs(in []multiaddr.Multiaddr) []net.IP {
	public := make([]net.IP, 0)
	for _, addr := range in {
		for _, code := range []int{multiaddr.P_IP4, multiaddr.P_IP6} {
			if v, err := addr.ValueForProtocol(code); err == nil {
				ip := net.ParseIP(v)
				if ip == nil {
					continue
				}
				if ip.IsPrivate() || ip.IsLoopback() {
					continue
				}
				public = append(public, ip)
			}
		}
	}
	return public
}
