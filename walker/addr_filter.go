package walker

import (
	"net"

	"github.com/multiformats/go-multiaddr"
)

type AddressFilter func(multiaddr.Multiaddr) bool

var DefaultAddressFilter = AddressFilterPublic

func AddressFilterAllowAll(multiaddr.Multiaddr) bool {
	return true
}

func AddressFilterPublic(addr multiaddr.Multiaddr) bool {
	for _, code := range []int{multiaddr.P_IP4, multiaddr.P_IP6} {
		if v, err := addr.ValueForProtocol(code); err == nil {
			ip := net.ParseIP(v)
			if ip == nil || ip.IsPrivate() || ip.IsLoopback() {
				return false
			}
			return true
		}
	}
	return false
}

func AddressFilterPrivate(addr multiaddr.Multiaddr) bool {
	return !AddressFilterPublic(addr)
}

func AddressFilterIpv6(addr multiaddr.Multiaddr) bool {
	for _, code := range []int{multiaddr.P_IP6} {
		if _, err := addr.ValueForProtocol(code); err == nil {
			return true
		}
	}
	return false
}

func AddressFilterIpv4(addr multiaddr.Multiaddr) bool {
	for _, code := range []int{multiaddr.P_IP4} {
		if _, err := addr.ValueForProtocol(code); err == nil {
			return true
		}
	}
	return false
}

func AddressFilterTcp(addr multiaddr.Multiaddr) bool {
	for _, code := range []int{multiaddr.P_TCP} {
		if _, err := addr.ValueForProtocol(code); err == nil {
			return true
		}
	}
	return false
}

func AddressFilterUdp(addr multiaddr.Multiaddr) bool {
	for _, code := range []int{multiaddr.P_UDP} {
		if _, err := addr.ValueForProtocol(code); err == nil {
			return true
		}
	}
	return false
}
