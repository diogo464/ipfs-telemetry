package telemetry

type Bandwidth struct {
	UploadRate   uint32 `json:"upload_rate"`
	DownloadRate uint32 `json:"download_rate"`
}
