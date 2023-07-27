package tupleconv_test

import (
	"github.com/google/uuid"
	dec "github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/go-tarantool/datetime"
	"github.com/tarantool/go-tarantool/decimal"
	"github.com/tarantool/go-tupleconv"
	"math/big"
	"testing"
	"time"
)

func TestStringToTTConvFactory(t *testing.T) {
	someUUID, err := uuid.Parse("09b56913-11f0-4fa4-b5d0-901b5efa532a")
	require.NoError(t, err)
	nullUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	time1, err := time.Parse(time.RFC3339, "2020-08-22T11:27:43.123456789-02:00")
	require.NoError(t, err)
	datetime1, err := datetime.NewDatetime(time1.UTC())
	require.NoError(t, err)

	time2, err := time.Parse(time.RFC3339, "1880-01-01T00:00:00Z")
	require.NoError(t, err)
	datetime2, err := datetime.NewDatetime(time2.UTC())
	require.NoError(t, err)

	time3, err := time.Parse("2006-01-02 15:04:05", "2023-08-30 11:11:00")
	require.NoError(t, err)
	datetime3, err := datetime.NewDatetime(time3.UTC())
	require.NoError(t, err)

	fac := tupleconv.MakeStringToTTConvFactory().
		WithNullValue("null").
		WithDecimalSeparators(",#").
		WithThousandSeparators(" `")

	convByType := map[tupleconv.TypeName]tupleconv.Converter[string, any]{
		tupleconv.TypeBoolean:   fac.GetBooleanConverter(),
		tupleconv.TypeInteger:   fac.GetIntegerConverter(),
		tupleconv.TypeUnsigned:  fac.GetUnsignedConverter(),
		tupleconv.TypeDouble:    fac.GetDoubleConverter(),
		tupleconv.TypeNumber:    fac.GetNumberConverter(),
		tupleconv.TypeDatetime:  fac.GetDatetimeConverter(),
		tupleconv.TypeUUID:      fac.GetUUIDConverter(),
		tupleconv.TypeArray:     fac.GetArrayConverter(),
		tupleconv.TypeVarbinary: fac.GetVarbinaryConverter(),
		tupleconv.TypeString:    fac.GetStringConverter(),
		tupleconv.TypeMap:       fac.GetMapConverter(),
		tupleconv.TypeAny:       fac.GetAnyConverter(),
		tupleconv.TypeScalar:    fac.GetScalarConverter(),
		tupleconv.TypeDecimal:   fac.GetDecimalConverter(),
		tupleconv.TypeInterval:  fac.GetIntervalConverter(),
	}

	tests := map[tupleconv.TypeName][]struct {
		value      string
		expected   any
		isNullable bool
		isErr      bool
	}{
		tupleconv.TypeBoolean: {
			// Basic.
			{value: "true", expected: true},
			{value: "false", expected: false},
			{value: "t", expected: true},
			{value: "f", expected: false},
			{value: "1", expected: true},
			{value: "0", expected: false},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "not bool at all", isErr: true},
			{value: "truth", isErr: true},
			{value: "`1", isErr: true},
		},
		tupleconv.TypeInteger: {
			// Basic.
			{value: "0", expected: uint64(0)},
			{value: "1", expected: uint64(1)},
			{value: "-1", expected: int64(-1)},
			{value: "12`13`144", expected: uint64(1213144)},
			{value: "-1`1", expected: int64(-11)},
			{value: "18446744073709551615", expected: uint64(18446744073709551615)},
			{value: "-9223372036854775808", expected: int64(-9223372036854775808)},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "1.2", isErr: true},
			{value: "-4329423948329482394328492349238", isErr: true}, // Too small.
			{value: "null", isErr: true},
			{value: "abacaba", isErr: true},
		},
		tupleconv.TypeUnsigned: {
			// Basic.
			{value: "1", expected: uint64(1)},
			{value: "18446744073709551615", expected: uint64(18446744073709551615)},
			{value: "0", expected: uint64(0)},
			{value: "12`13`144", expected: uint64(1213144)},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "18446744073709551616", isErr: true}, // Too big.
			{value: "null", isErr: true},
			{value: "111'111", isErr: true},
			{value: "str", isErr: true},
		},
		tupleconv.TypeDouble: {
			// Basic.
			{value: "12`13`144", expected: float64(1213144)},
			{value: "-11,12", expected: -11.12},
			{value: "0", expected: 0.0},
			{value: "1e-2", expected: 0.01},
			{value: "1.447e+44", expected: 1.447e+44},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "1'2'3*2", isErr: true},
			{value: "notdoubleatall", isErr: true},
			{value: `19.75z`, isErr: true},
		},
		tupleconv.TypeNumber: {
			// Basic.
			{value: "12`13`144", expected: uint64(1213144)},
			{value: "11,12", expected: 11.12},
			{value: "2.5", expected: 2.5},
			{value: "18446744073709551615", expected: uint64(18446744073709551615)},
			{value: "18446744073709551616", expected: float64(18446744073709551615)},
			{value: "0", expected: uint64(0)},
			{value: "-9223372036854775808", expected: int64(-9223372036854775808)},
			{value: "-9223372036854775809", expected: float64(-9223372036854775808)},
			{value: "439423943289", expected: uint64(439423943289)},
			{value: "1.15", expected: 1.15},
			{value: "1e-2", expected: 0.01},
			{value: "1.447e+44", expected: 1.447e+44},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "1'2'3'4**5", isErr: true},
			{value: "notnumberatall", isErr: true},
			{value: `{"a":3}`, isErr: true},
		},
		tupleconv.TypeDatetime: {
			// Basic.
			{value: "2020-08-22T11:27:43.123456789-02:00", expected: datetime1},
			{value: "1880-01-01T00:00:00Z", expected: datetime2},
			{value: "1880-01-01", expected: datetime2},
			{value: "2023-08-30 11:11:00", expected: datetime3},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "2020-08-22T11:27:43#123456789-02:00", isErr: true},
			{value: "19-19-19", isErr: true},
		},
		tupleconv.TypeUUID: {
			// Basic.
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532a", expected: someUUID},
			{value: "00000000-0000-0000-0000-000000000000", expected: nullUUID},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532", isErr: true},
		},
		tupleconv.TypeArray: {
			// Basic.
			{
				value:    "[1, 2, 3, 4]",
				expected: []any{float64(1), float64(2), float64(3), float64(4)},
			},
			{
				value: `[1, {"a" : [2,3]}]`,
				expected: []any{
					float64(1),
					map[string]any{
						"a": []any{float64(2), float64(3)},
					},
				},
			},
			{value: "null", expected: nil},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "[1,2,3,", isErr: true},
			{value: "[pqp][qpq]", isErr: true},
		},
		tupleconv.TypeVarbinary: {
			// Basic.
			{
				value:    "\x01\x02\x03",
				expected: []byte{1, 2, 3},
			},
			{
				value:    "abc",
				expected: []byte("abc"),
			},

			// Nullable.
			// Converting to null is prioritized over converting to base type.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, expected: []byte{}},
		},
		tupleconv.TypeString: {
			{value: "blablabla", expected: "blablabla"},
			{value: "бк#132433#$,%13п", expected: "бк#132433#$,%13п"},
			{value: "null", expected: "null"},

			// Nullable.
			// Converting to null is prioritized over converting to base type.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, expected: ""},
		},
		tupleconv.TypeMap: {
			// Basic.
			{
				value: `{"a":2, "b":3, "c":{ "d":"str" }}`,
				expected: map[string]any{
					"a": float64(2),
					"b": float64(3),
					"c": map[string]any{
						"d": "str",
					},
				},
			},
			{
				value: `{"1": [1,2,3], "2": {"a":4} }`,
				expected: map[string]any{
					"1": []any{float64(1), float64(2), float64(3)},
					"2": map[string]any{
						"a": float64(4),
					},
				},
			},

			// Nullable.
			// Parsing to the primary type is prioritized over null parsing.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: `{1:"2e5"}`, isErr: true},
			{value: `{"a":2 3}`, isErr: true},
			{value: `str`, isErr: true},
		},
		tupleconv.TypeAny: {
			{value: "1`e2", expected: 100.0},
			{value: "blablabla", expected: "blablabla"},
			{value: "0", expected: uint64(0)},
			{value: "1", expected: uint64(1)},
			{value: "-9223372036854775808", expected: int64(-9223372036854775808)},
			{value: "-9223372036854775809", expected: float64(-9223372036854775808)},
			{value: "true", expected: true},
			{value: "false", expected: false},
			{value: "", isNullable: true, expected: ""},
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532a", expected: someUUID},
			{value: "2020-08-22T11:27:43.123456789-02:00", expected: datetime1},
			{
				value: "1,2,3,4,5,6,7,8,1",
				expected: datetime.Interval{
					Year: 1, Month: 2, Week: 3, Day: 4, Hour: 5, Min: 6, Sec: 7, Nsec: 8, Adjust: 1,
				},
			},
			{
				value: "-11,-1110,0,0,0,0,0,1,2",
				expected: datetime.Interval{
					Year: -11, Month: -1110, Nsec: 1, Adjust: 2,
				},
			},
		},
		tupleconv.TypeScalar: {
			{value: "1`e2", expected: 100.0},
			{value: "blablabla", expected: "blablabla"},
			{value: "0", expected: uint64(0)},
			{value: "1", expected: uint64(1)},
			{value: "-9223372036854775808", expected: int64(-9223372036854775808)},
			{value: "-9223372036854775809", expected: float64(-9223372036854775808)},
			{value: "true", expected: true},
			{value: "false", expected: false},
			{value: "", isNullable: true, expected: ""},
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532a", expected: someUUID},
			{value: "2020-08-22T11:27:43.123456789-02:00", expected: datetime1},
			{
				value: "1,2,3,4,5,6,7,8,1",
				expected: datetime.Interval{
					Year: 1, Month: 2, Week: 3, Day: 4, Hour: 5, Min: 6, Sec: 7, Nsec: 8, Adjust: 1,
				},
			},
			{
				value: "-11,-1110,0,0,0,0,0,1,2",
				expected: datetime.Interval{
					Year: -11, Month: -1110, Nsec: 1, Adjust: 2,
				},
			},
		},
		tupleconv.TypeDecimal: {
			{
				value: "12`13`144",
				expected: &decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(1213144), 0),
				},
			},
			{
				value: "111`22e333",
				expected: &decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(11122), 333),
				},
			},
			{
				value: "11,12",
				expected: &decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(1112), -2),
				},
			},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "1,,2", isErr: true},
			{value: "hello", isErr: true},
		},
		tupleconv.TypeInterval: {
			// Basic.
			{
				value: "1,2,3,4,5,6,7,8,1",
				expected: datetime.Interval{
					Year: 1, Month: 2, Week: 3, Day: 4, Hour: 5, Min: 6, Sec: 7, Nsec: 8, Adjust: 1,
				},
			},
			{
				value: "-11,-1110,0,0,0,0,0,1,2",
				expected: datetime.Interval{
					Year: -11, Month: -1110, Nsec: 1, Adjust: 2,
				},
			},

			// Nullable.
			{value: "null", isNullable: true, expected: nil},
			{value: "", isNullable: true, isErr: true},

			// Error.
			{value: "1,2,3,4,5,6,7,8,3", isErr: true}, // Invalid adjust.
			{value: "1#2#3#4#5#6#7#8#3", isErr: true},
			{value: "1,2,3", isErr: true},
			{value: "1,2,3,4,5,6,7a,8,0", isErr: true},
			{value: "", isErr: true},
		},
	}

	for typ, cases := range tests {
		for _, tc := range cases {
			t.Run(string(typ)+" "+tc.value, func(t *testing.T) {
				converter := convByType[typ]
				assert.NoError(t, err)
				if tc.isNullable {
					converter = fac.MakeNullableConverter(converter)
				}
				converted, err := converter.Convert(tc.value)
				if tc.isErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.expected, converted)
				}
			})
		}
	}
}

type MockTypeToTTConvFactory struct{}

func (m MockTypeToTTConvFactory) GetBooleanConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "boolean", nil
	})
}

func (m MockTypeToTTConvFactory) GetStringConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "string", nil
	})
}

func (m MockTypeToTTConvFactory) GetUnsignedConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "unsigned", nil
	})
}

func (m MockTypeToTTConvFactory) GetDatetimeConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "datetime", nil
	})
}

func (m MockTypeToTTConvFactory) GetUUIDConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "uuid", nil
	})
}

func (m MockTypeToTTConvFactory) GetMapConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "map", nil
	})
}

func (m MockTypeToTTConvFactory) GetArrayConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "array", nil
	})
}

func (m MockTypeToTTConvFactory) GetVarbinaryConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "varbinary", nil
	})
}

func (m MockTypeToTTConvFactory) GetDoubleConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "double", nil
	})
}

func (m MockTypeToTTConvFactory) GetDecimalConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "decimal", nil
	})
}

func (m MockTypeToTTConvFactory) GetIntegerConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "integer", nil
	})
}

func (m MockTypeToTTConvFactory) GetNumberConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "number", nil
	})
}

func (m MockTypeToTTConvFactory) GetAnyConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "any", nil
	})
}

func (m MockTypeToTTConvFactory) GetScalarConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "scalar", nil
	})
}

func (m MockTypeToTTConvFactory) GetIntervalConverter() tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		return "interval", nil
	})
}

func (m MockTypeToTTConvFactory) MakeNullableConverter(
	c tupleconv.Converter[any, any]) tupleconv.Converter[any, any] {
	return tupleconv.MakeFuncConverter(func(s any) (any, error) {
		_, _ = c.Convert(s)
		return "null", nil
	})
}

var _ tupleconv.TTConvFactory[any] = (*MockTypeToTTConvFactory)(nil)

func TestGetConverterByType(t *testing.T) {
	fac := MockTypeToTTConvFactory{}
	types := [...]tupleconv.TypeName{
		tupleconv.TypeBoolean,
		tupleconv.TypeString,
		tupleconv.TypeInteger,
		tupleconv.TypeUnsigned,
		tupleconv.TypeDouble,
		tupleconv.TypeNumber,
		tupleconv.TypeDecimal,
		tupleconv.TypeDatetime,
		tupleconv.TypeUUID,
		tupleconv.TypeArray,
		tupleconv.TypeMap,
		tupleconv.TypeVarbinary,
		tupleconv.TypeScalar,
		tupleconv.TypeAny,
		tupleconv.TypeInterval,
	}
	for _, typ := range types {
		conv, err := tupleconv.GetConverterByType[any](fac, typ)
		assert.NoError(t, err)
		converted, _ := conv.Convert(nil)
		assert.Equal(t, string(typ), converted)
	}
	_, err := tupleconv.GetConverterByType[any](fac, "fake")
	assert.Error(t, err)
}

func TestMakeTypeToTTConverters_basic(t *testing.T) {
	spaceFmt := []tupleconv.SpaceField{
		{Type: "boolean"},
		{Type: "boolean", IsNullable: true},
		{Type: "integer"},
		{Type: "integer", IsNullable: true},
		{Type: "string"},
		{Type: "string", Name: "some name", IsNullable: true},
		{Type: "unsigned", Name: "some name"},
		{Type: "unsigned", IsNullable: true},
		{Type: "double"},
		{Type: "double", IsNullable: true},
		{Type: "number"},
		{Type: "datetime"},
		{Type: "datetime", IsNullable: true},
		{Type: "uuid"},
		{Type: "array"},
		{Type: "array", IsNullable: true},
		{Type: "varbinary", IsNullable: true},
		{Type: "map", Id: 1},
		{Type: "map", IsNullable: true},
		{Type: "any"},
		{Type: "scalar"},
		{Type: "scalar", IsNullable: true},
		{Type: "decimal"},
		{Type: "interval"},
	}

	fac := MockTypeToTTConvFactory{}
	converters, err := tupleconv.MakeTypeToTTConverters[any](fac, spaceFmt)
	assert.NoError(t, err)
	assert.Equal(t, len(spaceFmt), len(converters))
	for i, conv := range converters {
		converted, err := conv.Convert(nil)
		assert.NoError(t, err)
		if spaceFmt[i].IsNullable {
			assert.Equal(t, "null", converted)
		} else {
			assert.Equal(t, string(spaceFmt[i].Type), converted)
		}
	}
}

func TestMakeTypeToTTConverters_unexpected_type(t *testing.T) {
	spaceFmt := []tupleconv.SpaceField{
		{Type: "integer"},
		{Type: "fake", IsNullable: true},
	}
	fac := MockTypeToTTConvFactory{}
	_, err := tupleconv.MakeTypeToTTConverters[any](fac, spaceFmt)
	assert.Error(t, err)
}
