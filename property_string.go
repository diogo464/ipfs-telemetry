package telemetry

import (
	"context"
	"io"
)

var _ (PropertyCollector) = (*StringProperty)(nil)

type StringPropertyConfig struct {
	Name  string
	Value string
}

type StringProperty struct {
	descriptor PropertyDescriptor
	value      []byte
}

func NewStringProperty(config StringPropertyConfig) PropertyCollector {
	return &StringProperty{
		descriptor: PropertyDescriptor{
			Name:     config.Name,
			Encoding: EncodingString,
			Constant: true,
		},
		value: []byte(config.Value),
	}
}

// Collect implements Property
func (p *StringProperty) Collect(_ context.Context, w io.Writer) error {
	_, err := w.Write(p.value)
	return err
}

// Descriptor implements Property
func (p *StringProperty) Descriptor() PropertyDescriptor {
	return p.descriptor
}
