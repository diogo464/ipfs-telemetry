# Welcome to MkDocs

## Configuration

```go
type Telemetry struct {
	Enabled          bool
	BandwidthEnabled bool
	AccessType       telemetry.ServiceAccessType

	Whitelist            []peer.ID `json:",omitempty"`
	MetricsPeriod        string    `json:",omitempty"`
	WindowDuration       string    `json:",omitempty"`
	ActiveBufferDuration string    `json:",omitempty"`
	DebugListener        string    `json:",omitempty"`
}
```

`Enabled`
:   Enable or Disable the telemetry exporter

`BandwidthEnabled`
:   Enable or Disable bandwidth testing

`AccessType`
:   Control the peers that can request data.
Can be one of: `public`, `restricted`, `disabled`.  
`public` allows any peer to request data.  
`restricted` only allows peers in `Whitelist` to request data.  
`disabled` blocks every request.  
