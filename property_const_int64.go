package telemetry

import (
	"context"
	"encoding/binary"
	"io"
)

var _ (Property) = (*ConstInt64Property)(nil)

type ConstInt64Property struct {
	descriptor PropertyDescriptor
	value      int64
}

func NewConstInt64Property(name string, value int64) *ConstInt64Property {
	descriptor := PropertyDescriptor{
		Name:     name,
		Encoding: EncodingInt64,
	}
	return &ConstInt64Property{
		descriptor: descriptor,
		value:      value,
	}
}

// Collect implements Property
func (p *ConstInt64Property) Collect(_ context.Context, w io.Writer) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(p.value))
	_, err := w.Write(buf)
	return err
}

// Descriptor implements Property
func (p *ConstInt64Property) Descriptor() PropertyDescriptor {
	return p.descriptor
}
