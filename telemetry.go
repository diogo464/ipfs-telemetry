package telemetry

import (
	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("telemetry")

const (
	ID_TELEMETRY                   = "/telemetry/telemetry/0.4.0"
	ID_UPLOAD                      = "/telemetry/upload/0.4.0"
	ID_DOWNLOAD                    = "/telemetry/download/0.4.0"
	DEFAULT_BANDWIDTH_PAYLOAD_SIZE = 32 * 1024 * 1024
	MAX_BANDWIDTH_PAYLOAD_SIZE     = 128 * 1024 * 1024
	DATAPOINT_FETCH_BLOCK_SIZE     = 128
	DATAPOINT_UPLOAD_RATE          = 1024

	BLOCK_DURATION_BANDWIDTH = time.Minute * 5
	BLOCK_DURATION_STREAM    = time.Minute * 5

	METRICS_STREAM_ID = StreamId(0)
)
