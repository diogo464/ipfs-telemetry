package telemetry

import (
	"context"
	"fmt"
	"io"
)

var ErrPropertyAlreadyRegistered = fmt.Errorf("property already registered")

type PropertyOption func(*propertyConfig) error

type Property interface {
	Descriptor() PropertyDescriptor
	// Write the property value to the writer.
	// Must be thread safe.
	Collect(context.Context, io.Writer) error
}

type PropertyDescriptor struct {
	Name     string
	Encoding string
}

type propertyConfig struct {
	overrideName     *string
	overrideEncoding *string
}

func propertyConfigDefaults() *propertyConfig {
	return &propertyConfig{
		overrideName:     nil,
		overrideEncoding: nil,
	}
}

func propertyConfigApply(c *propertyConfig, opts ...PropertyOption) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

func WithPropertyOverrideName(name string) PropertyOption {
	return func(c *propertyConfig) error {
		c.overrideName = &name
		return nil
	}
}

func WithPropertyOverrideEncoding(encoding string) PropertyOption {
	return func(c *propertyConfig) error {
		c.overrideEncoding = &encoding
		return nil
	}
}
