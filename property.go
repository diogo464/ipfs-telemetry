package telemetry

import "github.com/diogo464/telemetry/internal/pb"

func propertyConfigToPb(c PropertyConfig) *pb.Property {
	p := &pb.Property{
		Name:        c.Name,
		Description: c.Description,
	}

	switch c.Value.(type) {
	case *PropertyValueInteger:
		p.Value = &pb.Property_IntegerValue{
			IntegerValue: c.Value.GetInteger(),
		}
	case *PropertyValueString:
		p.Value = &pb.Property_StringValue{
			StringValue: c.Value.GetString(),
		}
	default:
		panic("not implemented")
	}

	return p
}

func propertyPbToClientProperty(c *pb.Property) CProperty {
	p := CProperty{
		Name:        c.GetName(),
		Description: c.GetDescription(),
		Value:       nil,
	}

	switch v := c.Value.(type) {
	case *pb.Property_IntegerValue:
		p.Value = NewPropertyValueInteger(v.IntegerValue)
	case *pb.Property_StringValue:
		p.Value = NewPropertyValueString(v.StringValue)
	}

	return p
}
