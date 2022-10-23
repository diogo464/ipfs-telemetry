package telemetry

import (
	"context"
	"io"
)

var _ (Property) = (*ConstStrProperty)(nil)

type ConstStrProperty struct {
	descriptor PropertyDescriptor
	value      []byte
}

func NewConstStrProperty(name string, value string) *ConstStrProperty {
	descriptor := PropertyDescriptor{
		Name:     name,
		Encoding: EncodingString,
	}
	return &ConstStrProperty{
		descriptor: descriptor,
		value:      []byte(value),
	}
}

// Collect implements Property
func (p *ConstStrProperty) Collect(_ context.Context, w io.Writer) error {
	_, err := w.Write(p.value)
	return err
}

// Descriptor implements Property
func (p *ConstStrProperty) Descriptor() PropertyDescriptor {
	return p.descriptor
}
