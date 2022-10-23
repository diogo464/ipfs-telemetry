package telemetry

import (
	"time"

	logging "github.com/ipfs/go-log"
)

// log is the command logger
var log = logging.Logger("telemetry")

type Encoding uint32

const (
	EncodingBinary  = 0
	EncodingInt64   = 1
	EncodingFloat64 = 2
	EncodingString  = 3
	EncodingJson    = 4
)

const (
	ID_TELEMETRY               = "/telemetry/telemetry/0.4.0"
	ID_UPLOAD                  = "/telemetry/upload/0.4.0"
	ID_DOWNLOAD                = "/telemetry/download/0.4.0"
	DEFAULT_PAYLOAD_SIZE       = 32 * 1024 * 1024
	MAX_PAYLOAD_SIZE           = 128 * 1024 * 1024
	DATAPOINT_FETCH_BLOCK_SIZE = 128
	DATAPOINT_UPLOAD_RATE      = 1024

	BLOCK_DURATION_BANDWIDTH = time.Minute * 5
	BLOCK_DURATION_STREAM    = time.Minute * 5
)

func ReadableEncoding(e Encoding) string {
	switch e {
	case EncodingBinary:
		return "binary"
	case EncodingInt64:
		return "int64"
	case EncodingFloat64:
		return "float64"
	case EncodingString:
		return "string"
	case EncodingJson:
		return "json"
	default:
		return "unknown"
	}
}

func TimestampNow() uint64 {
	return uint64(time.Now().UnixNano())
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
