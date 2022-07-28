package telemetry

import (
	"context"
	"encoding/json"
	"io"
)

var _ Property = (*simpleProperty)(nil)
var _ Property = (*jsonProperty)(nil)

type SimplePropertyDescriptor struct {
	Name     string
	Encoding string
	Collect  func(context.Context, io.Writer) error
}

type simpleProperty struct {
	name     string
	encoding string
	collect  func(context.Context, io.Writer) error
}

// Collect implements Property
func (p *simpleProperty) Collect(ctx context.Context, writer io.Writer) error {
	return p.collect(ctx, writer)
}

// Descriptor implements Property
func (p *simpleProperty) Descriptor() PropertyDescriptor {
	return PropertyDescriptor{
		Name:     p.name,
		Encoding: p.encoding,
	}
}

func SimpleProperty(d SimplePropertyDescriptor) Property {
	return &simpleProperty{
		name:     d.Name,
		encoding: d.Encoding,
		collect:  d.Collect,
	}
}

type JsonPropertyDescriptor struct {
	Name    string
	Collect func(context.Context) (interface{}, error)
}

type jsonProperty struct {
	name    string
	collect func(context.Context) (interface{}, error)
}

// Collect implements Property
func (p *jsonProperty) Collect(ctx context.Context, writer io.Writer) error {
	obj, err := p.collect(ctx)
	if err != nil {
		return err
	}
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = writer.Write(marshaled)
	return err
}

// Descriptor implements Property
func (p *jsonProperty) Descriptor() PropertyDescriptor {
	return PropertyDescriptor{
		Name:     p.name,
		Encoding: ENCODING_JSON,
	}
}

func JsonProperty(d JsonPropertyDescriptor) Property {
	return &jsonProperty{
		name:    d.Name,
		collect: d.Collect,
	}
}
