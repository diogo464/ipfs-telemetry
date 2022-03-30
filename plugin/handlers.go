package telemetry

import (
	"context"
	"fmt"
	"io"
	"log"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"git.d464.sh/adc/telemetry/plugin/utils"
	"git.d464.sh/adc/telemetry/plugin/wire"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/network"
)

func (t *TelemetryService) TelemetryHandler(s network.Stream) {
	defer s.Close()

	request, err := wire.ReadRequest(context.TODO(), s)
	if err != nil {
		return
	}

	switch request.GetBody().(type) {
	case *pb.Request_Snapshots_:
		err = t.handleRequestSnapshots(s, request)
	case *pb.Request_SystemInfo_:
		err = t.handleRequestSystemInfo(s, request)
	case *pb.Request_BandwidthDownload_:
		err = t.handleRequestBandwidthDownload(s, request)
	case *pb.Request_BandwidthUpload_:
		err = t.handleRequestBandwidthUpload(s, request)
	default:
	}

	fmt.Println("done handling request")

	if err != nil {
		fmt.Println(err)
	}
}

func (t *TelemetryService) handleRequestSnapshots(s network.Stream, request *pb.Request) error {
	fmt.Println("handlign request snapshots")
	var since uint64 = 0
	requests_snapshots := request.GetSnapshots()
	fmt.Println("snapshots obtained")
	remote_session, err := uuid.Parse(requests_snapshots.Session)
	if err != nil {
		return err
	}
	if remote_session == t.s {
		since = requests_snapshots.Since
	}
	fmt.Println("writting response")
	return wire.WriteResponse(context.TODO(), s, wire.NewSnapshots(t.s, t.w.Since(since)))
}

func (t *TelemetryService) handleRequestSystemInfo(s network.Stream, request *pb.Request) error {
	return wire.WriteResponse(context.TODO(), s, wire.NewSystemInfo(t.s))
}

func (t *TelemetryService) handleRequestBandwidthDownload(s network.Stream, request *pb.Request) error {
	log.Println("Handle BandwidthDownload")
	io.Copy(io.Discard, io.LimitReader(s, BANDWIDTH_PAYLOAD_SIZE))
	return nil
}

func (t *TelemetryService) handleRequestBandwidthUpload(s network.Stream, request *pb.Request) error {
	log.Println("Handle BandwidthUpload")
	io.Copy(s, io.LimitReader(utils.NullReader{}, BANDWIDTH_PAYLOAD_SIZE))
	return nil
}
