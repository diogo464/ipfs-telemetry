package telemetry

import (
	"time"

	logging "github.com/ipfs/go-log"
)

// log is the command logger
var log = logging.Logger("telemetry")

const (
	ID_TELEMETRY               = "/telemetry/telemetry/0.2.0"
	ID_UPLOAD                  = "/telemetry/upload/0.2.0"
	ID_DOWNLOAD                = "/telemetry/download/0.2.0"
	DEFAULT_PAYLOAD_SIZE       = 32 * 1024 * 1024
	MAX_PAYLOAD_SIZE           = 128 * 1024 * 1024
	DATAPOINT_FETCH_BLOCK_SIZE = 128
	DATAPOINT_UPLOAD_RATE      = 1024

	BLOCK_DURATION_BANDWIDTH = time.Minute * 5
	BLOCK_DURATION_STREAM    = time.Minute * 5

	ENCODING_JSON     = "json"
	ENCODING_PROTOBUF = "protobuf"
	ENCODING_CUSTOM   = "custom"
	ENCODING_UNKNOWN  = "unknown"
)

type SystemInfo struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	NumCPU uint32 `json:"numcpu"`
}

type SessionInfo struct {
	Session  Session   `json:"session"`
	BootTime time.Time `json:"boottime"`
}

type Bandwidth struct {
	UploadRate   uint32 `json:"upload_rate"`
	DownloadRate uint32 `json:"download_rate"`
}

type DebugStream struct {
	Name      string `json:"name"`
	UsedSize  uint32 `json:"used_size"`
	TotalSize uint32 `json:"total_size"`
}

type Debug struct {
	Streams []DebugStream
}
