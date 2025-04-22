package telemetry

import (
	"context"
	"fmt"
	"net"

	"github.com/diogo464/telemetry/internal/utils"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	grpc_peer "google.golang.org/grpc/peer"
)

func getPublicIpFromContext(h host.Host, ctx context.Context) (net.IP, error) {
	grpcPeer, ok := grpc_peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to obtain peer")
	}
	// https://github.com/libp2p/go-libp2p-gostream/blob/master/addr.go
	pidB58 := grpcPeer.Addr.String()
	pid, err := peer.Decode(pidB58)
	if err != nil {
		return nil, err
	}
	addrs := h.Peerstore().Addrs(pid)
	return utils.GetFirstPublicAddressFromMultiaddrs(addrs)
}
