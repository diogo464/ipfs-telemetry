package telemetry

import (
	"strconv"

	"github.com/diogo464/telemetry/internal/pb"
)

var (
	_ (PropertyValue) = (*propertyValueInteger)(nil)
	_ (PropertyValue) = (*propertyValueString)(nil)
)

type PropertyValue interface {
	sealed()

	GetString() string
	GetInteger() int64

	String() string
}

type PropertyDescriptor struct {
	ID          uint32
	Scope       string
	Name        string
	Description string
}

type Property struct {
	Scope       string
	Name        string
	Description string
	Value       PropertyValue
}

type propertyValueString struct {
	value string
}

type propertyValueInteger struct {
	value int64
}

type propertyConfig struct {
	Scope       string
	Name        string
	Description string
	// Value is one of PropertyValueInteger, PropertyValueString
	Value PropertyValue
}

func PropertyValueString(v string) PropertyValue {
	return &propertyValueString{value: v}
}

func PropertyValueInteger(v int64) PropertyValue {
	return &propertyValueInteger{value: v}
}

func propertyConfigToPb(id uint32, c propertyConfig) *pb.Property {
	p := &pb.Property{
		Id:          id,
		Scope:       c.Scope,
		Name:        c.Name,
		Description: c.Description,
	}

	switch c.Value.(type) {
	case *propertyValueInteger:
		p.Value = &pb.Property_IntegerValue{
			IntegerValue: c.Value.GetInteger(),
		}
	case *propertyValueString:
		p.Value = &pb.Property_StringValue{
			StringValue: c.Value.GetString(),
		}
	default:
		panic("not implemented")
	}

	return p
}

func propertyPbToClientProperty(c *pb.Property) Property {
	p := Property{
		Scope:       c.GetScope(),
		Name:        c.GetName(),
		Description: c.GetDescription(),
		Value:       nil,
	}

	switch v := c.Value.(type) {
	case *pb.Property_IntegerValue:
		p.Value = PropertyValueInteger(v.IntegerValue)
	case *pb.Property_StringValue:
		p.Value = PropertyValueString(v.StringValue)
	}

	return p
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
