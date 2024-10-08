package tupleconv_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	dec "github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/go-tarantool/v2/datetime"
	"github.com/tarantool/go-tarantool/v2/decimal"
	"github.com/tarantool/go-tupleconv"
)

type convCase[S any, T any] struct {
	value    S
	expected T
	isErr    bool
}

func HelperTestConverter[S any, T any](
	t *testing.T,
	mp tupleconv.Converter[S, T],
	cases []convCase[S, T]) {
	for _, tc := range cases {
		t.Run(fmt.Sprintln(tc.value), func(t *testing.T) {
			result, err := mp.Convert(tc.value)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func getDatetimeWithValidate(t *testing.T, tm time.Time) datetime.Datetime {
	dt, err := datetime.MakeDatetime(tm)
	require.NoError(t, err)
	return dt
}

func TestConverters(t *testing.T) {
	someUUID, err := uuid.Parse("09b56913-11f0-4fa4-b5d0-901b5efa532a")
	require.NoError(t, err)
	nullUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	parisLoc, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	time1 := time.Date(2020, 8, 22, 11, 27, 43,
		123456789, time.FixedZone("", -2*60*60))
	time2 := time.Date(2022, 7, 26, 17, 6, 49, 809892000,
		time.FixedZone("", +3*60*60))
	time3 := time.Date(2022, 7, 26, 17, 06, 49,
		809892000, parisLoc)
	time4 := time.Date(2023, 8, 30, 12, 6, 5, 120000000,
		time.FixedZone("", 0))
	time5 := time.Date(2023, 8, 30, 12, 6, 5, 120000000,
		parisLoc)

	datetime1 := getDatetimeWithValidate(t, time1)
	datetime2 := getDatetimeWithValidate(t, time2)
	datetime3 := getDatetimeWithValidate(t, time3)
	datetime4 := getDatetimeWithValidate(t, time4)
	datetime5 := getDatetimeWithValidate(t, time5)

	thSeparators := "@ '"
	decSeparators := ".*"

	tests := map[tupleconv.Converter[string, any]][]convCase[string, any]{
		tupleconv.MakeStringToBoolConverter(): {
			// Basic.
			{value: "true", expected: true},
			{value: "false", expected: false},
			{value: "t", expected: true},
			{value: "f", expected: false},
			{value: "1", expected: true},
			{value: "0", expected: false},

			// Error.
			{value: "not bool at all", isErr: true},
			{value: "truth", isErr: true},
		},
		tupleconv.MakeStringToUIntConverter(thSeparators): {
			// Basic.
			{value: "1", expected: uint64(1)},
			{value: "18446744073709551615", expected: uint64(18446744073709551615)},
			{value: "0", expected: uint64(0)},
			{value: "439423943289", expected: uint64(439423943289)},
			{value: "111'111'111'111", expected: uint64(111111111111)},

			// Error.
			{value: "18446744073709551616", isErr: true}, // Too big.
			{value: "-101010", isErr: true},
			{value: "null", isErr: true},
			{value: "111`111", isErr: true},
			{value: "str", isErr: true},
		},
		tupleconv.MakeStringToIntConverter(thSeparators): {
			// Basic.
			{value: "0", expected: int64(0)},
			{value: "1", expected: int64(1)},
			{value: "111'111'111'111", expected: int64(111111111111)},
			{value: "111@111@111'111", expected: int64(111111111111)},
			{value: "-1", expected: int64(-1)},
			{value: "9223372036854775807", expected: int64(9223372036854775807)},
			{value: "-9223372036854775808", expected: int64(-9223372036854775808)},
			{value: "-115 92 28 239", expected: int64(-1159228239)},

			// Error.
			{value: "-9223372036854775809", isErr: true}, // Too small.
			{value: "9223372036854775808", isErr: true},  // Too big.
			{value: "null", isErr: true},
			{value: "14,15", isErr: true},
			{value: "2.5", isErr: true},
			{value: "abacaba", isErr: true},
		},
		tupleconv.MakeStringToFloatConverter(thSeparators, decSeparators): {
			// Basic.
			{value: "1.15", expected: 1.15},
			{value: "1e-2", expected: 0.01},
			{value: "-44", expected: float64(-44)},
			{value: "1.447e+44", expected: 1.447e+44},
			{value: "1", expected: float64(1)},
			{value: "-1", expected: float64(-1)},
			{value: "18446744073709551615", expected: float64(18446744073709551615)},
			{value: "18446744073709551616", expected: float64(18446744073709551615)},
			{value: "0", expected: float64(0)},
			{value: "-9223372036854775808", expected: float64(-9223372036854775808)},
			{value: "-9223372036854775809", expected: float64(-9223372036854775808)},
			{value: "439423943289", expected: float64(439423943289)},
			{value: "1.15", expected: 1.15},
			{value: "1e-2", expected: 0.01},
			{value: "1.447e+44", expected: 1.447e+44},
			{value: "1 2 3 @ 4", expected: float64(1234)},
			{value: "1 2 3 * 4", expected: 123.4},

			// Error.
			{value: "1'2'3'4**5", isErr: true},
			{value: "notnumberatall", isErr: true},
			{value: `{"a":3}`, isErr: true},
		},
		tupleconv.MakeStringToDatetimeConverter(): {
			// Basic.
			{value: "2020-08-22T11:27:43.123456789-0200", expected: datetime1},
			{value: "2022-07-26T17:06:49.809892+0300", expected: datetime2},
			{value: "2022-07-26T17:06:49.809892 Europe/Paris", expected: datetime3},
			{value: "2023-08-30T12:06:05.120-0000", expected: datetime4},
			{value: "2023-08-30T12:06:05.120 Europe/Paris", expected: datetime5},

			// Error.
			{value: "19-19-19", isErr: true},
			{value: "#$,%13п", isErr: true},
			{value: "2020-08-22T11:27:43*123456789-02:00", isErr: true},
			{value: "2023-08-30T12:06:05.120 Tatuin", isErr: true},
			{value: "2023-08-30 Europe/Moscow", isErr: true},
		},
		tupleconv.MakeStringToUUIDConverter(): {
			// Basic.
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532a", expected: someUUID},
			{value: "00000000-0000-0000-0000-000000000000", expected: nullUUID},

			// Error.
			{value: "09b56913-11f0-4fa4-b5d0-901b5efa532", isErr: true},
		},
		tupleconv.MakeStringToSliceConverter(): {
			// Basic.
			{
				value:    "[1, 2, 3, 4]",
				expected: []any{float64(1), float64(2), float64(3), float64(4)}},
			{
				value: `[1, {"a" : [2,3]}]`,
				expected: []any{
					float64(1),
					map[string]any{
						"a": []any{float64(2), float64(3)},
					},
				},
			},
			{
				value:    "[null, null, null]",
				expected: []any{nil, nil, nil},
			},

			// Error.
			{value: "[1,2,3,", isErr: true},
			{value: "[pqp][qpq]", isErr: true},
		},
		tupleconv.MakeStringToMapConverter(): {
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
			{
				value: `{"1" : null, "2" : null}`,
				expected: map[string]any{
					"1": nil,
					"2": nil,
				},
			},

			// Error.
			{value: `{1:"2"}`, isErr: true},
			{value: `{"a":2`, isErr: true},
			{value: `str`, isErr: true},
		},
		tupleconv.MakeStringToBinaryConverter(): {
			// Basic.
			{value: "\x01\x02\x03", expected: []byte{1, 2, 3}},
			{value: "abc", expected: []byte("abc")},
		},
		tupleconv.MakeIdentityConverter[string](): {
			// Basic.
			{value: "blablabla", expected: "blablabla"},
			{value: "бк#132433#$,%13п", expected: "бк#132433#$,%13п"},
			{value: "null", expected: "null"},
		},
		tupleconv.MakeStringToNullConverter("null"): {
			// Basic.
			{value: "null", expected: nil},

			// Error.
			{value: "505", isErr: true},
			{value: "nil", isErr: true},
		},
		tupleconv.MakeStringToDecimalConverter(thSeparators, decSeparators): {
			{
				value:    "0",
				expected: decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(0), 0)},
			},
			{
				value:    "1",
				expected: decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(1), 0)},
			},
			{
				value:    "-1",
				expected: decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(-1), 0)},
			},
			{
				value: "43904329",
				expected: decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(43904329), 0),
				},
			},
			{
				value: "-9223372036854775808",
				expected: decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(int64(-9223372036854775808)), 0),
				},
			},
			{
				value:    "1.447e+44",
				expected: decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(1447), 41)},
			},
			{
				value:    "1*5",
				expected: decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(15), -1)},
			},

			// Error.
			{value: "abacaba", isErr: true},
			{value: "-1/0", isErr: true},
			{value: "1**5", isErr: true},
		},
		tupleconv.MakeStringToIntervalConverter(): {
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

			// Error.
			{value: "1,2,3,4,5,6,7,8,3", isErr: true}, // Invalid adjust.
			{value: "1,2,3", isErr: true},
			{value: "1,2,3,4,5,6,7a,8,0", isErr: true},
			{value: "", isErr: true},
		},
	}

	for parser, cases := range tests {
		HelperTestConverter(t, parser, cases)
	}
}

func TestMakeDatetimeToStringConverter(t *testing.T) {
	parisLoc, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	time1 := time.Date(2023, 8, 30, 12, 6, 5, 123456789,
		time.FixedZone("", 60*60*4))
	time2 := time.Date(2023, 8, 30, 12, 6, 5, 120000000,
		parisLoc)
	time3 := time.Date(2020, 9, 14, 12, 12, 12, 0, time.UTC)

	cases := []convCase[datetime.Datetime, string]{
		{
			value:    getDatetimeWithValidate(t, time1),
			expected: "2023-08-30T12:06:05.123456789+0400",
		},
		{
			value:    getDatetimeWithValidate(t, time2),
			expected: "2023-08-30T12:06:05.12 Europe/Paris",
		},
		{
			value:    getDatetimeWithValidate(t, time3),
			expected: "2020-09-14T12:12:12 UTC",
		},
	}
	converter := tupleconv.MakeDatetimeToStringConverter()
	HelperTestConverter[datetime.Datetime, string](t, converter, cases)
}

func TestMakeIntervalToStringConverter(t *testing.T) {
	interval := datetime.Interval{
		Year:   1234,
		Month:  11,
		Week:   22,
		Day:    33,
		Hour:   44,
		Min:    55,
		Sec:    66,
		Nsec:   77,
		Adjust: datetime.LastAdjust,
	}

	sparseInterval := datetime.Interval{
		Year: 1414,
		Nsec: 4343904394,
	}

	cases := []convCase[datetime.Interval, string]{
		{value: interval, expected: "1234,11,22,33,44,55,66,77,2"},
		{value: sparseInterval, expected: "1414,0,0,0,0,0,0,4343904394,0"},
	}
	converter := tupleconv.MakeIntervalToStringConverter()
	HelperTestConverter[datetime.Interval, string](t, converter, cases)
}

func TestMakeSequenceConverter(t *testing.T) {
	parser := tupleconv.MakeSequenceConverter([]tupleconv.Converter[string, any]{
		tupleconv.MakeStringToUIntConverter(""),
		tupleconv.MakeStringToIntConverter(""),
		tupleconv.MakeStringToFloatConverter("", "."),
		tupleconv.MakeStringToMapConverter(),
	})

	cases := []convCase[string, any]{
		// Basic.
		{value: "0", expected: uint64(0)},
		{value: "1", expected: uint64(1)},
		{value: "-10", expected: int64(-10)},
		{value: "2.5", expected: 2.5},
		{value: "{}", expected: map[string]any{}},
		{value: "null", expected: nil}, // As `json`.

		// Error.
		{value: "12-13-14", isErr: true},
		{value: "12,14", isErr: true},
	}
	HelperTestConverter(t, parser, cases)
}
