package telemetry

import (
	"strconv"

	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	_ (PropertyValue) = (*PropertyValueInteger)(nil)
	_ (PropertyValue) = (*PropertyValueString)(nil)
)

type Property struct {
	Scope       instrumentation.Scope `json:"scope"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Value       PropertyValue         `json:"value"`
}

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

func NewPropertyValueString(v string) PropertyValue {
	return &PropertyValueString{value: v}
}

func NewPropertyValueInteger(v int64) PropertyValue {
	return &PropertyValueInteger{value: v}
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
