package tupleconv

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool/datetime"
	"github.com/tarantool/go-tarantool/decimal"
)

// Converter is a converter from S to T.
type Converter[S any, T any] interface {
	Convert(src S) (T, error)
}

// Interface validations.
var (
	_ Converter[string, any] = (*StringToBoolConverter)(nil)
	_ Converter[string, any] = (*StringToUIntConverter)(nil)
	_ Converter[string, any] = (*StringToIntConverter)(nil)
	_ Converter[string, any] = (*StringToFloatConverter)(nil)
	_ Converter[string, any] = (*StringToDecimalConverter)(nil)
	_ Converter[string, any] = (*StringToUUIDConverter)(nil)
	_ Converter[string, any] = (*StringToDatetimeConverter)(nil)
	_ Converter[string, any] = (*StringToMapConverter)(nil)
	_ Converter[string, any] = (*StringToSliceConverter)(nil)
	_ Converter[string, any] = (*StringToNullConverter)(nil)
	_ Converter[string, any] = (*IdentityConverter[string])(nil)
	_ Converter[string, any] = (*StringToIntervalConverter)(nil)
)

// IdentityConverter is a converter from S to any, that doesn't change the input.
type IdentityConverter[S any] struct{}

// MakeIdentityConverter creates IdentityConverter.
func MakeIdentityConverter[S any]() IdentityConverter[S] {
	return IdentityConverter[S]{}
}

// Convert is the implementation of Converter[S, any] for IdentityConverter.
func (IdentityConverter[S]) Convert(src S) (any, error) {
	return src, nil
}

// FuncConverter is a function-based Converter.
type FuncConverter[S any, T any] struct {
	convFunc func(S) (T, error)
}

// MakeFuncConverter creates FuncConverter.
func MakeFuncConverter[S any, T any](convFunc func(S) (T, error)) FuncConverter[S, T] {
	return FuncConverter[S, T]{convFunc: convFunc}
}

// Convert is the implementation of Converter for FuncConverter.
func (conv FuncConverter[S, T]) Convert(src S) (T, error) {
	return conv.convFunc(src)
}

// MakeSequenceConverter makes a sequential Converter from a Converter list.
func MakeSequenceConverter[S any, T any](converters []Converter[S, T]) Converter[S, T] {
	return MakeFuncConverter(func(src S) (T, error) {
		for _, conv := range converters {
			if result, err := conv.Convert(src); err == nil {
				return result, nil
			}
		}
		var ret T
		return ret, fmt.Errorf("unexpected value %v", src)
	})
}

// replaceSeparators replaces all characters from `charsToReplace` with a specific string.
func replaceCharacters(src, charsToReplace, replaceTo string) string {
	for _, char := range charsToReplace {
		src = strings.ReplaceAll(src, string(char), replaceTo)
	}
	return src
}

// StringToBoolConverter is a converter from string to bool.
type StringToBoolConverter struct{}

// MakeStringToBoolConverter creates StringToBoolConverter.
func MakeStringToBoolConverter() StringToBoolConverter {
	return StringToBoolConverter{}
}

// Convert is the implementation of Converter[string, any] for StringToBoolConverter.
func (StringToBoolConverter) Convert(src string) (any, error) {
	return strconv.ParseBool(src)
}

// StringToUIntConverter is a converter from string to uint64.
type StringToUIntConverter struct {
	ignoreChars string
}

// MakeStringToUIntConverter creates StringToUIntConverter.
func MakeStringToUIntConverter(ignoreChars string) StringToUIntConverter {
	return StringToUIntConverter{ignoreChars: ignoreChars}
}

// Convert is the implementation of Converter[string, any] for StringToUIntConverter.
func (conv StringToUIntConverter) Convert(src string) (any, error) {
	src = replaceCharacters(src, conv.ignoreChars, "")
	return strconv.ParseUint(src, 10, 64)
}

// StringToIntConverter is a converter from string to int64.
type StringToIntConverter struct {
	ignoreChars string
}

// MakeStringToIntConverter creates StringToIntConverter.
func MakeStringToIntConverter(ignoreChars string) StringToIntConverter {
	return StringToIntConverter{ignoreChars: ignoreChars}
}

// Convert is the implementation of Converter[string, any] for StringToIntConverter.
func (conv StringToIntConverter) Convert(src string) (any, error) {
	src = replaceCharacters(src, conv.ignoreChars, "")
	return strconv.ParseInt(src, 10, 64)
}

// StringToFloatConverter is a converter from string to float64.
type StringToFloatConverter struct {
	ignoreChars   string
	decSeparators string
}

// MakeStringToFloatConverter creates StringToFloatConverter.
func MakeStringToFloatConverter(ignoreChars, decSeparators string) StringToFloatConverter {
	return StringToFloatConverter{ignoreChars: ignoreChars, decSeparators: decSeparators}
}

// Convert is the implementation of Converter[string, any] for StringToFloatConverter.
func (conv StringToFloatConverter) Convert(src string) (any, error) {
	src = replaceCharacters(src, conv.ignoreChars, "")
	src = replaceCharacters(src, conv.decSeparators, ".")
	return strconv.ParseFloat(src, 64)
}

// StringToDecimalConverter is a converter from string to decimal.Decimal.
type StringToDecimalConverter struct {
	ignoreChars   string
	decSeparators string
}

// MakeStringToDecimalConverter creates StringToDecimalConverter.
func MakeStringToDecimalConverter(ignoreChars, decSeparators string) StringToDecimalConverter {
	return StringToDecimalConverter{ignoreChars: ignoreChars, decSeparators: decSeparators}
}

// Convert is the implementation of Converter[string, any] for StringToDecimalConverter.
func (conv StringToDecimalConverter) Convert(src string) (any, error) {
	src = replaceCharacters(src, conv.ignoreChars, "")
	src = replaceCharacters(src, conv.decSeparators, ".")
	return decimal.NewDecimalFromString(src)
}

// StringToUUIDConverter is a converter from string to UUID.
type StringToUUIDConverter struct{}

// MakeStringToUUIDConverter creates StringToUUIDConverter.
func MakeStringToUUIDConverter() StringToUUIDConverter {
	return StringToUUIDConverter{}
}

// Convert is the implementation of Converter[string, any] for StringToUUIDConverter.
func (StringToUUIDConverter) Convert(src string) (any, error) {
	return uuid.Parse(src)
}

// StringToDatetimeConverter is a converter from string to datetime.Datetime.
// Time layouts used:
// - RFC3339
// - 2006-01-02 15:04:05
// - 2006-01-02
// - 2006-01-02T15:04:05.999999-0700.
type StringToDatetimeConverter struct{}

// MakeStringToDatetimeConverter creates StringToDatetimeConverter.
func MakeStringToDatetimeConverter() StringToDatetimeConverter {
	return StringToDatetimeConverter{}
}

const (
	ttTimeLayout   = "2006-01-02T15:04:05.999999-0700"
	dateTimeLayout = "2006-01-02 15:04:05"
	dateOnlyLayout = "2006-01-02"
)

// Convert is the implementation of Converter[string, any] for StringToDatetimeConverter.
func (StringToDatetimeConverter) Convert(src string) (any, error) {
	for _, layout := range [...]string{
		time.RFC3339,
		dateTimeLayout,
		dateOnlyLayout,
		ttTimeLayout,
	} {
		result, err := time.Parse(layout, src)
		if err != nil {
			continue
		}
		return datetime.NewDatetime(result.UTC())
	}
	return nil, fmt.Errorf("unexpected value: %v", src)
}

// StringToMapConverter is a converter from string to map.
// Only `json` is supported now.
type StringToMapConverter struct{}

// MakeStringToMapConverter creates StringToMapConverter.
func MakeStringToMapConverter() StringToMapConverter {
	return StringToMapConverter{}
}

// Convert is the implementation of Converter[string, any] for StringToMapConverter.
func (StringToMapConverter) Convert(src string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(src), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StringToSliceConverter is a converter from string to slice.
// Only `json` is supported now.
type StringToSliceConverter struct{}

// MakeStringToSliceConverter creates StringToSliceConverter.
func MakeStringToSliceConverter() StringToSliceConverter {
	return StringToSliceConverter{}
}

// Convert is the implementation of Converter[string, any] for StringToSliceConverter.
func (StringToSliceConverter) Convert(src string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(src), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StringToBinaryConverter is a converter from string to binary.
type StringToBinaryConverter struct{}

// MakeStringToBinaryConverter creates StringToBinaryConverter.
func MakeStringToBinaryConverter() StringToBinaryConverter {
	return StringToBinaryConverter{}
}

// Convert is the implementation of Converter[string, any] for StringToBinaryConverter.
func (StringToBinaryConverter) Convert(src string) (any, error) {
	return []byte(src), nil
}

// StringToNullConverter is a converter from string to nil.
type StringToNullConverter struct {
	nullValue string
}

// MakeStringToNullConverter creates StringToNullConverter.
func MakeStringToNullConverter(nullValue string) StringToNullConverter {
	return StringToNullConverter{nullValue: nullValue}
}

// Convert is the implementation of Converter[string, any] for StringToNullConverter.
func (conv StringToNullConverter) Convert(src string) (any, error) {
	if src != conv.nullValue {
		return nil, fmt.Errorf("unexpected value: %v", src)
	}
	return nil, nil
}

// StringToIntervalConverter is a converter from string to datetime.Interval.
type StringToIntervalConverter struct{}

// MakeStringToIntervalConverter creates StringToIntervalConverter.
func MakeStringToIntervalConverter() StringToIntervalConverter {
	return StringToIntervalConverter{}
}

// intervalFieldsNumber is the number of fields in datetime.Interval.
const intervalFieldsNumber = 9

var errUnexpectedIntervalFmt = errors.New("unexpected interval format")

// Convert is the implementation of Converter[string, any] for StringToIntervalConverter.
func (StringToIntervalConverter) Convert(src string) (any, error) {
	parts := strings.Split(src, ",")
	if len(parts) != intervalFieldsNumber {
		return nil, errUnexpectedIntervalFmt
	}
	partsAsInt64 := [intervalFieldsNumber]int64{}
	for i, part := range parts {
		var err error
		if partsAsInt64[i], err = strconv.ParseInt(part, 10, 64); err != nil {
			return nil, errUnexpectedIntervalFmt
		}
	}
	adjust := datetime.Adjust(partsAsInt64[8])
	if adjust != datetime.NoneAdjust &&
		adjust != datetime.ExcessAdjust && adjust != datetime.LastAdjust {
		return nil, errUnexpectedIntervalFmt
	}
	interval := datetime.Interval{
		Year:   partsAsInt64[0],
		Month:  partsAsInt64[1],
		Week:   partsAsInt64[2],
		Day:    partsAsInt64[3],
		Hour:   partsAsInt64[4],
		Min:    partsAsInt64[5],
		Sec:    partsAsInt64[6],
		Nsec:   partsAsInt64[7],
		Adjust: adjust,
	}
	return interval, nil
}
