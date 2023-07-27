package tupleconv

import (
	"fmt"
)

// TypeName is the data type for type names.
type TypeName string

// Types are supported tarantool types.
const (
	TypeBoolean   TypeName = "boolean"
	TypeString    TypeName = "string"
	TypeInteger   TypeName = "integer"
	TypeUnsigned  TypeName = "unsigned"
	TypeDouble    TypeName = "double"
	TypeNumber    TypeName = "number"
	TypeDecimal   TypeName = "decimal"
	TypeDatetime  TypeName = "datetime"
	TypeUUID      TypeName = "uuid"
	TypeArray     TypeName = "array"
	TypeMap       TypeName = "map"
	TypeVarbinary TypeName = "varbinary"
	TypeScalar    TypeName = "scalar"
	TypeAny       TypeName = "any"
	TypeInterval  TypeName = "interval"
)

const (
	defaultThousandSeparators = ""
	defaultDecimalSeparators  = "."
	defaultNullValue          = ""
)

// TTConvFactory is a factory capable of creating converters from Type
// to tarantool types.
type TTConvFactory[Type any] interface {
	// GetBooleanConverter returns a converter from Type to boolean.
	GetBooleanConverter() Converter[Type, any]

	// GetStringConverter returns a converter from Type to string.
	GetStringConverter() Converter[Type, any]

	// GetUnsignedConverter returns a converter from Type to unsigned.
	GetUnsignedConverter() Converter[Type, any]

	// GetDatetimeConverter returns a converter from Type to datetime.
	GetDatetimeConverter() Converter[Type, any]

	// GetUUIDConverter returns a converter from Type to uuid.
	GetUUIDConverter() Converter[Type, any]

	// GetMapConverter returns a converter from Type to map.
	GetMapConverter() Converter[Type, any]

	// GetArrayConverter returns a converter from Type to array.
	GetArrayConverter() Converter[Type, any]

	// GetVarbinaryConverter returns a converter from Type to varbinary.
	GetVarbinaryConverter() Converter[Type, any]

	// GetDoubleConverter returns a converter from Type to double.
	GetDoubleConverter() Converter[Type, any]

	// GetDecimalConverter returns a converter from Type to decimal.
	GetDecimalConverter() Converter[Type, any]

	// GetIntegerConverter returns a converter from Type to integer.
	GetIntegerConverter() Converter[Type, any]

	// GetNumberConverter returns a converter from Type to number.
	GetNumberConverter() Converter[Type, any]

	// GetAnyConverter returns a converter from Type to any.
	GetAnyConverter() Converter[Type, any]

	// GetScalarConverter returns a converter from Type to scalar.
	GetScalarConverter() Converter[Type, any]

	// GetIntervalConverter returns a converter from Type to interval.
	GetIntervalConverter() Converter[Type, any]

	// MakeNullableConverter extends the incoming converter to a nullable converter.
	MakeNullableConverter(Converter[Type, any]) Converter[Type, any]
}

// StringToTTConvFactory is the default TypeToTTConvFactory for strings.
// To customize the creation of converters, inherit from it and override the necessary methods.
type StringToTTConvFactory struct {
	// thousandSeparators are thousands separators for numeric types.
	// In fact, these symbols will be ignored for numeric types.
	// For example:
	// thousandSeparators = "`#e": 1`2`3 -> 123, 123#456#789 -> 123456789, 1e5 -> 15.
	thousandSeparators string

	// decimalSeparators are additional decimal separators for numeric types (besides `.`).
	// In fact, these symbols will be replaced with `.`.
	// For example:
	// decimalSeparators = ",#": 12,3 -> 12.3, 100.500 -> 100.500, 12#13 -> 12.13.
	decimalSeparators string

	// nullValue is a value that is interpreted as null.
	nullValue string
}

// MakeStringToTTConvFactory creates StringToTTConvFactory.
func MakeStringToTTConvFactory() StringToTTConvFactory {
	return StringToTTConvFactory{
		thousandSeparators: defaultThousandSeparators,
		decimalSeparators:  defaultDecimalSeparators,
		nullValue:          defaultNullValue,
	}
}

func (StringToTTConvFactory) GetBooleanConverter() Converter[string, any] {
	return MakeStringToBoolConverter()
}

func (StringToTTConvFactory) GetStringConverter() Converter[string, any] {
	return MakeIdentityConverter[string]()
}

func (fac StringToTTConvFactory) GetUnsignedConverter() Converter[string, any] {
	return MakeStringToUIntConverter(fac.thousandSeparators)
}

func (StringToTTConvFactory) GetDatetimeConverter() Converter[string, any] {
	return MakeStringToDatetimeConverter()
}

func (StringToTTConvFactory) GetUUIDConverter() Converter[string, any] {
	return MakeStringToUUIDConverter()
}

func (StringToTTConvFactory) GetMapConverter() Converter[string, any] {
	return MakeStringToMapConverter()
}

func (StringToTTConvFactory) GetArrayConverter() Converter[string, any] {
	return MakeStringToSliceConverter()
}

func (StringToTTConvFactory) GetVarbinaryConverter() Converter[string, any] {
	return MakeStringToBinaryConverter()
}

func (fac StringToTTConvFactory) GetDoubleConverter() Converter[string, any] {
	return MakeStringToFloatConverter(fac.thousandSeparators, fac.decimalSeparators)
}

func (fac StringToTTConvFactory) GetDecimalConverter() Converter[string, any] {
	return MakeStringToDecimalConverter(fac.thousandSeparators, fac.decimalSeparators)
}

func (fac StringToTTConvFactory) GetIntegerConverter() Converter[string, any] {
	return MakeSequenceConverter([]Converter[string, any]{
		MakeStringToUIntConverter(fac.thousandSeparators),
		MakeStringToIntConverter(fac.thousandSeparators),
	})
}

func (fac StringToTTConvFactory) GetNumberConverter() Converter[string, any] {
	return MakeSequenceConverter([]Converter[string, any]{
		MakeStringToUIntConverter(fac.thousandSeparators),
		MakeStringToIntConverter(fac.thousandSeparators),
		MakeStringToFloatConverter(fac.thousandSeparators, fac.decimalSeparators),
	})
}

func (fac StringToTTConvFactory) GetIntervalConverter() Converter[string, any] {
	return MakeStringToIntervalConverter()
}

func (fac StringToTTConvFactory) GetAnyConverter() Converter[string, any] {
	return MakeSequenceConverter([]Converter[string, any]{
		fac.GetNumberConverter(),
		fac.GetDecimalConverter(),
		fac.GetBooleanConverter(),
		fac.GetDatetimeConverter(),
		fac.GetUUIDConverter(),
		fac.GetIntervalConverter(),
		fac.GetStringConverter(),
	})
}

func (fac StringToTTConvFactory) GetScalarConverter() Converter[string, any] {
	return MakeSequenceConverter([]Converter[string, any]{
		fac.GetNumberConverter(),
		fac.GetDecimalConverter(),
		fac.GetBooleanConverter(),
		fac.GetDatetimeConverter(),
		fac.GetUUIDConverter(),
		fac.GetIntervalConverter(),
		fac.GetStringConverter(),
	})
}

func (fac StringToTTConvFactory) MakeNullableConverter(
	converter Converter[string, any]) Converter[string, any] {
	return MakeSequenceConverter([]Converter[string, any]{
		MakeStringToNullConverter(fac.nullValue),
		converter,
	})
}

// WithNullValue sets nullValue.
func (fac StringToTTConvFactory) WithNullValue(nullValue string) StringToTTConvFactory {
	fac.nullValue = nullValue
	return fac
}

// WithThousandSeparators sets thousandSeparators.
func (fac StringToTTConvFactory) WithThousandSeparators(
	separators string) StringToTTConvFactory {
	fac.thousandSeparators = separators
	return fac
}

// WithDecimalSeparators sets decimalSeparators.
func (fac StringToTTConvFactory) WithDecimalSeparators(separators string) StringToTTConvFactory {
	fac.decimalSeparators = separators
	return fac
}

var _ TTConvFactory[string] = (*StringToTTConvFactory)(nil)

// GetConverterByType returns a converter by TTConvFactory and typename.
func GetConverterByType[Type any](
	fac TTConvFactory[Type], typ TypeName) (conv Converter[Type, any], err error) {
	switch typ {
	case TypeBoolean:
		conv = fac.GetBooleanConverter()
	case TypeString:
		conv = fac.GetStringConverter()
	case TypeUnsigned:
		conv = fac.GetUnsignedConverter()
	case TypeDatetime:
		conv = fac.GetDatetimeConverter()
	case TypeUUID:
		conv = fac.GetUUIDConverter()
	case TypeMap:
		conv = fac.GetMapConverter()
	case TypeArray:
		conv = fac.GetArrayConverter()
	case TypeVarbinary:
		conv = fac.GetVarbinaryConverter()
	case TypeDouble:
		conv = fac.GetDoubleConverter()
	case TypeDecimal:
		conv = fac.GetDecimalConverter()
	case TypeInteger:
		conv = fac.GetIntegerConverter()
	case TypeNumber:
		conv = fac.GetNumberConverter()
	case TypeAny:
		conv = fac.GetAnyConverter()
	case TypeScalar:
		conv = fac.GetScalarConverter()
	case TypeInterval:
		conv = fac.GetIntervalConverter()
	default:
		return nil, fmt.Errorf("unexpected type: %s", typ)
	}
	return
}

// SpaceField is a space field.
type SpaceField struct {
	Id         uint32   `msgpack:"id,omitempty"`
	Name       string   `msgpack:"name"`
	Type       TypeName `msgpack:"type"`
	IsNullable bool     `msgpack:"is_nullable,omitempty"`
}

// MakeTypeToTTConverters creates list of the converters
// from Type to tt type by the factory and space format.
func MakeTypeToTTConverters[Type any](
	fac TTConvFactory[Type],
	spaceFmt []SpaceField) ([]Converter[Type, any], error) {
	converters := make([]Converter[Type, any], len(spaceFmt))
	for i, fieldFmt := range spaceFmt {
		typ := fieldFmt.Type
		conv, err := GetConverterByType(fac, typ)
		if err != nil {
			return nil, err
		}
		if fieldFmt.IsNullable {
			conv = fac.MakeNullableConverter(conv)
		}
		converters[i] = MakeFuncConverter(func(s Type) (any, error) {
			result, err := conv.Convert(s)
			if err != nil {
				return nil, fmt.Errorf("unexpected value %v for type %q", s, typ)
			}
			return result, nil
		})
	}
	return converters, nil
}
