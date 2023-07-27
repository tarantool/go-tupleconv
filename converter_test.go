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
	"github.com/tarantool/go-tarantool/datetime"
	"github.com/tarantool/go-tarantool/decimal"
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

func TestConverters(t *testing.T) {
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

	time4, err := time.Parse(
		"2006-01-02T15:04:05.999999-0700", "2022-07-26T17:06:49.809892+0300")
	require.NoError(t, err)
	datetime4, err := datetime.NewDatetime(time4.UTC())
	require.NoError(t, err)

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
			{value: "2020-08-22T11:27:43.123456789-02:00", expected: datetime1},
			{value: "1880-01-01T00:00:00Z", expected: datetime2},
			{value: "1880-01-01", expected: datetime2},
			{value: "2023-08-30 11:11:00", expected: datetime3},
			{value: "2022-07-26T17:06:49.809892+0300", expected: datetime4},

			// Error.
			{value: "19-19-19", isErr: true},
			{value: "#$,%13п", isErr: true},
			{value: "2020-08-22T11:27:43*123456789-02:00", isErr: true},
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
				expected: &decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(0), 0)},
			},
			{
				value:    "1",
				expected: &decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(1), 0)},
			},
			{
				value:    "-1",
				expected: &decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(-1), 0)},
			},
			{
				value: "43904329",
				expected: &decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(43904329), 0),
				},
			},
			{
				value: "-9223372036854775808",
				expected: &decimal.Decimal{
					Decimal: dec.NewFromBigInt(big.NewInt(int64(-9223372036854775808)), 0),
				},
			},
			{
				value:    "1.447e+44",
				expected: &decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(1447), 41)},
			},
			{
				value:    "1*5",
				expected: &decimal.Decimal{Decimal: dec.NewFromBigInt(big.NewInt(15), -1)},
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
