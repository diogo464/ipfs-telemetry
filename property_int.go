package telemetry

import (
	"context"
	"encoding/binary"
	"io"
)

var _ (PropertyCollector) = (*IntProperty)(nil)

type IntPropertyConfig struct {
	Name  string
	Value int64
}

type IntProperty struct {
	descriptor PropertyDescriptor
	value      int64
}

func NewIntProperty(config IntPropertyConfig) PropertyCollector {
	descriptor := PropertyDescriptor{
		Name:     config.Name,
		Encoding: EncodingInt64,
		Constant: true,
	}
	return &IntProperty{
		descriptor: descriptor,
		value:      config.Value,
	}
}

// Collect implements Property
func (p *IntProperty) Collect(_ context.Context, w io.Writer) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(p.value))
	_, err := w.Write(buf)
	return err
}

// Descriptor implements Property
func (p *IntProperty) Descriptor() PropertyDescriptor {
	return p.descriptor
}
