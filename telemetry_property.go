package telemetry

import (
	"strconv"

	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	_ (PropertyValue) = (*propertyValueInteger)(nil)
	_ (PropertyValue) = (*propertyValueString)(nil)
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

type propertyValueString struct {
	value string
}

type propertyValueInteger struct {
	value int64
}

func PropertyValueString(v string) PropertyValue {
	return &propertyValueString{value: v}
}

func PropertyValueInteger(v int64) PropertyValue {
	return &propertyValueInteger{value: v}
}

// GetInteger implements PropertyValue
func (p *propertyValueInteger) GetInteger() int64 {
	return p.value
}

// GetString implements PropertyValue
func (p *propertyValueInteger) GetString() string {
	return ""
}

// sealed implements PropertyValue
func (*propertyValueInteger) sealed() {
}

// String implements PropertyValue
func (p *propertyValueInteger) String() string {
	return strconv.Itoa(int(p.value))
}

// GetInteger implements PropertyValue
func (*propertyValueString) GetInteger() int64 {
	return 0
}

// GetString implements PropertyValue
func (p *propertyValueString) GetString() string {
	return p.value
}

// sealed implements PropertyValue
func (*propertyValueString) sealed() {
}

// String implements PropertyValue
func (p *propertyValueString) String() string {
	return p.value
}
