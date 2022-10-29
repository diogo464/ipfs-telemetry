package telemetry

import (
	"context"
	"strconv"
	"time"

	logging "github.com/ipfs/go-log"
	"go.opentelemetry.io/otel/metric"
)

// log is the command logger
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
)

var (
	_ (PropertyValue) = (*PropertyValueInteger)(nil)
	_ (PropertyValue) = (*PropertyValueString)(nil)
)

type CaptureCallback func(context.Context) (interface{}, error)

type PropertyValue interface {
	sealed()

	GetString() string
	GetInteger() int64

	String() string
}

type PropertyValueString struct {
	value string
}

type PropertyValueInteger struct {
	value int64
}

type PropertyConfig struct {
	Name        string
	Description string
	// Value is one of PropertyValueInteger, PropertyValueString
	Value PropertyValue
}

type CaptureConfig struct {
	Name        string
	Description string
	Callback    CaptureCallback
	Interval    time.Duration
}

type CaptureDescriptor struct {
	Name        string
	Description string
}

type EventConfig struct {
	Name        string
	Description string
}

type EventDescriptor struct {
	Name        string
	Description string
}

type EventEmitter interface {
	Emit(interface{})
}

type Telemetry interface {
	metric.MeterProvider

	Property(PropertyConfig)
	Capture(CaptureConfig)
	Event(EventConfig) EventEmitter
	// TODO: Add RPC's
}

func TimestampNow() uint64 {
	return uint64(time.Now().UnixNano())
}

type Bandwidth struct {
	UploadRate   uint32 `json:"upload_rate"`
	DownloadRate uint32 `json:"download_rate"`
}

func NewPropertyValueInteger(value int64) *PropertyValueInteger {
	return &PropertyValueInteger{
		value: value,
	}
}

func NewPropertyValueString(value string) *PropertyValueString {
	return &PropertyValueString{
		value: value,
	}
}

// GetInteger implements PropertyValue
func (p *PropertyValueInteger) GetInteger() int64 {
	return p.value
}

// GetString implements PropertyValue
func (p *PropertyValueInteger) GetString() string {
	return ""
}

// sealed implements PropertyValue
func (*PropertyValueInteger) sealed() {
}

// String implements PropertyValue
func (p *PropertyValueInteger) String() string {
	return strconv.Itoa(int(p.value))
}

// GetInteger implements PropertyValue
func (*PropertyValueString) GetInteger() int64 {
	return 0
}

// GetString implements PropertyValue
func (p *PropertyValueString) GetString() string {
	return p.value
}

// sealed implements PropertyValue
func (*PropertyValueString) sealed() {
}

// String implements PropertyValue
func (p *PropertyValueString) String() string {
	return p.value
}
