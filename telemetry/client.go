package telemetry

import (
	"context"
	"fmt"
	"io"
	"time"

	"git.d464.sh/adc/telemetry/telemetry/utils"
	"git.d464.sh/adc/telemetry/telemetry/wire"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type TelemetryClient struct {
	h host.Host
	p peer.ID
}

func NewTelemetryClient(h host.Host, p peer.ID) *TelemetryClient {
	return &TelemetryClient{h: h, p: p}
}

func (c *TelemetryClient) Snapshots(ctx context.Context, session uuid.UUID, since uint64) (*wire.ResponseSnapshot, error) {
	request := wire.NewRequestSnapshot(session, since)
	response, err := c.makeRequest(ctx, request, wire.RESPONSE_SNAPSHOT)
	if err != nil {
		return nil, err
	}
	return response.GetSnapshot()
}

func (c *TelemetryClient) SystemInfo(ctx context.Context) (*wire.ResponseSystemInfo, error) {
	request := wire.NewRequestSystemInfo()
	response, err := c.makeRequest(ctx, request, wire.RESPONSE_SYSTEM_INFO)
	if err != nil {
		return nil, err
	}
	return response.GetSystemInfo()
}

func (c *TelemetryClient) Download(ctx context.Context) (uint64, error) {
	s, err := c.h.NewStream(ctx, c.p, ID)
	if err != nil {
		return 0, err
	}
	defer s.Close()

	request := wire.NewRequestBandwdithDownload()
	if err := wire.WriteRequest(ctx, s, request); err != nil {
		return 0, err
	}

	write_start := time.Now()
	n, err := io.Copy(s, io.LimitReader(utils.NullReader{}, BANDWIDTH_PAYLOAD_SIZE))
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1)
	s.Read(buf)
	elapsed := time.Since(write_start)

	rate := uint64(float64(n) / elapsed.Seconds())
	return rate, nil
}

func (c *TelemetryClient) Upload(ctx context.Context) (uint64, error) {
	s, err := c.h.NewStream(ctx, c.p, ID)
	if err != nil {
		return 0, err
	}
	defer s.Close()

	request := wire.NewRequestBandwdithUpload()
	if err := wire.WriteRequest(ctx, s, request); err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024)
	n := 0
	var read_start *time.Time = nil
	for {
		x, err := s.Read(buf)
		n += x
		if read_start == nil {
			tm := time.Now()
			read_start = &tm
		}
		if err != nil && err != io.EOF {
			return 0, err
		}
		if err == io.EOF {
			break
		}
	}
	//n, err := io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
	if err != nil {
		return 0, err
	}
	elapsed := time.Since(*read_start)

	rate := uint64(float64(n) / elapsed.Seconds())
	return rate, nil
}

func (c *TelemetryClient) makeRequest(ctx context.Context, req *wire.Request, resp wire.ResponseType) (*wire.Response, error) {
	s, err := c.h.NewStream(ctx, c.p, ID)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	err = wire.WriteRequest(ctx, s, req)
	if err != nil {
		return nil, err
	}

	response, err := wire.ReadResponse(ctx, s)
	if err != nil {
		return nil, err
	}

	if response.Type != resp {
		return nil, fmt.Errorf("expected response %v got %v", resp, response.Type)
	}

	return response, nil
}
