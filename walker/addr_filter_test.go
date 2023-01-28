package walker_test

import (
	"testing"

	"github.com/diogo464/telemetry/walker"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func parseMultiAddr(addr string) multiaddr.Multiaddr {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		panic(err)
	}
	return maddr
}

func TestAddressFilterPublic(t *testing.T) {
	assert.True(t, walker.AddressFilterPublic(parseMultiAddr("/ip4/80.0.1.2/tcp/1234")))
	assert.True(t, walker.AddressFilterPublic(parseMultiAddr("/ip6/2001:db8::/tcp/1234")))
	assert.False(t, walker.AddressFilterPublic(parseMultiAddr("/ip4/127.0.0.1/tcp/1234")))
	assert.False(t, walker.AddressFilterPublic(parseMultiAddr("/ip4/127.0.0.1/udp/1234")))
	assert.False(t, walker.AddressFilterPublic(parseMultiAddr("/ip6/::1/tcp/1234")))
	assert.False(t, walker.AddressFilterPublic(parseMultiAddr("/ip6/::1/udp/1234")))
	assert.False(t, walker.AddressFilterPublic(parseMultiAddr("/ip4/10.5.0.1/tcp/1234")))
}

func TestAddressFilterIpv4(t *testing.T) {
	assert.True(t, walker.AddressFilterIpv4(parseMultiAddr("/ip4/10.0.0.1/tcp/1234")))
	assert.False(t, walker.AddressFilterIpv4(parseMultiAddr("/ip6/::1/tcp/1234")))
}

func TestAddressFilterIpv6(t *testing.T) {
	assert.True(t, walker.AddressFilterIpv6(parseMultiAddr("/ip6/::1/tcp/1234")))
	assert.False(t, walker.AddressFilterIpv6(parseMultiAddr("/ip4/80.0.0.1/tcp/1234")))
}

func TestAddressFilterTcp(t *testing.T) {
	assert.True(t, walker.AddressFilterTcp(parseMultiAddr("/ip4/10.0.1.1/tcp/1234")))
	assert.False(t, walker.AddressFilterTcp(parseMultiAddr("/ip4/10.0.1.1/udp/1234")))
}

func TestAddressFilterUdp(t *testing.T) {
	assert.False(t, walker.AddressFilterTcp(parseMultiAddr("/ip4/10.0.1.1/tcp/1234")))
	assert.True(t, walker.AddressFilterTcp(parseMultiAddr("/ip4/10.0.1.1/udp/1234")))
}
