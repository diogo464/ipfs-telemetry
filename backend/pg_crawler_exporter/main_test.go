package pg_crawler_exporter_test

import (
	"net"
	"slices"
	"testing"

	"github.com/diogo464/ipfs-telemetry/backend/pg_crawler_exporter"
	"github.com/multiformats/go-multiaddr"
)

func CreateSampleAddresses() []multiaddr.Multiaddr {
	addrStrings := []string{
		"/ip4/132.226.169.184/tcp/4001/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/35.79.127.140/udp/4001/quic",
		"/ip4/5.166.238.76/udp/50336/quic-v1/webtransport/certhash/uEiAuVXxdGPSOrIiBTrHiwwt0xKQg9jICllwdo_RGx5FPiw/certhash/uEiChf3IVddFEZtPWxsjVOLMyD4jE9gwFjTTgVAr2V-7ydQ/p2p/12D3KooWSVfiLWLjmVMGrm8QABTVoh79E8xWGN7BuLoif3bDQ9tt/p2p-circuit",
		"/ip6/64:ff9b::234f:7f8c/udp/4001/quic-v1/webtransport/certhash/uEiBbHljKvqSp2dcDTRWH9Nbyb6ariy1FbZyF2z9iFUPORA/certhash/uEiCs0BB3jRz7-VqQwomju1vP8FRCb6gOpfl7k8SAbgKAhA",
		"/ip6/64:ff9b::234f:7f8c/udp/4001/quic-v1",
		"/ip4/35.79.127.140/tcp/4001",
		"/ip4/5.166.238.76/udp/50336/quic-v1/p2p/12D3KooWSVfiLWLjmVMGrm8QABTVoh79E8xWGN7BuLoif3bDQ9tt/p2p-circuit",
		"/ip4/35.79.127.140/udp/4001/quic-v1/webtransport/certhash/uEiBbHljKvqSp2dcDTRWH9Nbyb6ariy1FbZyF2z9iFUPORA/certhash/uEiCs0BB3jRz7-VqQwomju1vP8FRCb6gOpfl7k8SAbgKAhA",
		"/ip4/118.33.40.93/udp/4001/quic-v1/webtransport/certhash/uEiBwnR07HvXax_Eu_W-lOVav4BMjoHTWgKRBRxgTc5QVGw/certhash/uEiDWs0DMVoQs4rxRfkRE8-pseUP42Qy1Q_RGexYTNPxNGA/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/118.33.40.93/udp/4001/quic-v1/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/118.33.40.93/udp/4001/quic/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/5.166.238.76/tcp/50336/p2p/12D3KooWSVfiLWLjmVMGrm8QABTVoh79E8xWGN7BuLoif3bDQ9tt/p2p-circuit",
		"/ip4/132.226.169.184/udp/4001/quic-v1/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/35.79.127.140/udp/4001/quic-v1",
		"/ip4/132.226.169.184/udp/4001/quic-v1/webtransport/certhash/uEiBwnR07HvXax_Eu_W-lOVav4BMjoHTWgKRBRxgTc5QVGw/certhash/uEiDWs0DMVoQs4rxRfkRE8-pseUP42Qy1Q_RGexYTNPxNGA/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
		"/ip4/118.33.40.93/tcp/4001/p2p/12D3KooWMeRfth7YpbCmy89YyMRmgjXRuCW9QzjvjuG6JSfoaqHC/p2p-circuit",
	}

	addrs := make([]multiaddr.Multiaddr, len(addrStrings))
	for i, s := range addrStrings {
		addr, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		addrs[i] = addr
	}
	return addrs
}

func TestExtractFirstPublicIp(t *testing.T) {
	ip, ok := pg_crawler_exporter.ExtractFirstPublicIp(CreateSampleAddresses())
	if !ok {
		t.Fatalf("failed to find ip")
	}

	if slices.Compare(ip, net.ParseIP("35.79.127.140")) != 0 {
		t.Fatalf("invalid ip %v", ip)
	}
}
