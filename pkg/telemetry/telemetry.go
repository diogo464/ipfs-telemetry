package telemetry

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	ID_TELEMETRY               = "/telemetry/telemetry/0.0.0"
	ID_UPLOAD                  = "/telemetry/upload/0.0.0"
	ID_DOWNLOAD                = "/telemetry/download/0.0.0"
	DEFAULT_PAYLOAD_SIZE       = 32 * 1024 * 1024
	MAX_PAYLOAD_SIZE           = 128 * 1024 * 1024
	DATAPOINT_FETCH_BLOCK_SIZE = 128
	DATAPOINT_UPLOAD_RATE      = 1024

	BLOCK_DURATION_BANDWIDTH          = time.Minute * 5
	BLOCK_DURATION_GETRECORDPROVIDERS = time.Minute * 5
	BLOCK_DURATION_GETDATAPOINTS      = time.Minute * 5
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

type ProviderRecordEntry struct {
	Peer        peer.ID   `json:"peer"`
	LastRefresh time.Time `json:"last_refresh"`
}

type ProviderRecord struct {
	Key     []byte                `json:"key"`
	Entries []ProviderRecordEntry `json:"entries"`
}
