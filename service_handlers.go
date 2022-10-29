package telemetry

import (
	"io"
	"time"

	"github.com/diogo464/telemetry/internal/utils"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

func (s *Service) uploadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.uploadBlocker.isBlocked(publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.uploadBlocker.block(publicIp, BLOCK_DURATION_BANDWIDTH)
	}

	requested_payload, err := utils.ReadU32(stream)
	if err != nil || requested_payload > MAX_BANDWIDTH_PAYLOAD_SIZE {
		return
	}

	upload_start := time.Now()
	n, err := io.Copy(stream, io.LimitReader(utils.NullReader{}, int64(requested_payload)))
	if err != nil {
		return
	}
	elapsed := time.Since(upload_start)
	if err != nil {
		return
	}
	rate := uint32(float64(n) / elapsed.Seconds())
	_ = utils.WriteU32(stream, rate)
}

func (s *Service) downloadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.downloadBlocker.isBlocked(publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.downloadBlocker.block(publicIp, BLOCK_DURATION_BANDWIDTH)
	}

	expected_payload, err := utils.ReadU32(stream)
	if err != nil || expected_payload > MAX_BANDWIDTH_PAYLOAD_SIZE {
		return
	}

	download_start := time.Now()
	n, err := io.Copy(io.Discard, io.LimitReader(stream, int64(expected_payload)))
	elapsed := time.Since(download_start)
	if err != nil {
		return
	}
	rate := uint32(float64(n) / elapsed.Seconds())
	_ = utils.WriteU32(stream, rate)
}
