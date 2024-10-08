package tupleconv_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
	"github.com/tarantool/go-tarantool/v2/test_helpers"
	"github.com/tarantool/go-tupleconv"

	_ "github.com/tarantool/go-tarantool/v2/uuid"
)

type filterIntConverter struct {
	toFilter string
}

func (c filterIntConverter) Convert(src string) (int64, error) {
	src = strings.ReplaceAll(src, c.toFilter, "")
	return strconv.ParseInt(src, 10, 64)
}

// ExampleConverter demonstrates the basic usage of the Converter.
func ExampleConverter() {
	// Basic converter.
	strToBoolConv := tupleconv.MakeStringToBoolConverter()
	result, err := strToBoolConv.Convert("true")
	fmt.Println(result, err)

	// Function based converter.
	funcConv := tupleconv.MakeFuncConverter(func(s string) (string, error) {
		return s + " world!", nil
	})
	result, err = funcConv.Convert("hello")
	fmt.Println(result, err)

	var filterConv tupleconv.Converter[string, int64] = filterIntConverter{toFilter: "th"}
	result, err = filterConv.Convert("100th")
	fmt.Println(result, err)

	// Output:
	// true <nil>
	// hello world! <nil>
	// 100 <nil>
}

// ExampleMapper_basicMapper demonstrates the basic usage of the Mapper.
func ExampleMapper_basicMapper() {
	// Mapper example.
	mapper := tupleconv.MakeMapper[string, any]([]tupleconv.Converter[string, any]{
		tupleconv.MakeFuncConverter(func(s string) (any, error) {
			return s + "1", nil
		}),
		tupleconv.MakeFuncConverter(func(s string) (any, error) {
			iVal, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.New("can't convert")
			}
			return iVal + 1, nil
		}),
	})
	result, err := mapper.Map([]string{"a", "4"})
	fmt.Println(result, err)

	result, err = mapper.Map([]string{"0"})
	fmt.Println(result, err)

	// Output:
	// [a1 5] <nil>
	// [01] <nil>
}

// ExampleMapper_singleMapper demonstrates the usage of the Mapper with
// only the default converter.
func ExampleMapper_singleMapper() {
	// Single mapper example.
	toStringMapper := tupleconv.MakeMapper([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(tupleconv.MakeFuncConverter(
			func(s any) (string, error) {
				return fmt.Sprint(s), nil
			}),
		)
	res, err := toStringMapper.Map([]any{1, 2.5, nil})
	fmt.Println(res, err)

	// Output:
	// [1 2.5 <nil>] <nil>
}

// ExampleStringToTTConvFactory demonstrates how to create Converter list for
// Mapper using helper functions and StringToTTConvFactory.
func ExampleStringToTTConvFactory() {
	factory := tupleconv.MakeStringToTTConvFactory().
		WithDecimalSeparators(",.")

	spaceFmt := []tupleconv.SpaceField{
		{Type: tupleconv.TypeUnsigned},
		{Type: tupleconv.TypeDouble, IsNullable: true},
		{Type: tupleconv.TypeString},
	}

	converters, _ := tupleconv.MakeTypeToTTConverters[string](factory, spaceFmt)
	mapper := tupleconv.MakeMapper(converters)
	result, err := mapper.Map([]string{"1", "-2,2", "some_string"})
	fmt.Println(result, err)

	// Output:
	// [1 -2.2 some_string] <nil>
}

// ExampleStringToTTConvFactory_manualConverters demonstrates how to obtain Converter
// from TTConvFactory for manual Converter list construction.
func ExampleStringToTTConvFactory_manualConverters() {
	factory := tupleconv.MakeStringToTTConvFactory().
		WithDecimalSeparators(",.")

	fieldTypes := []tupleconv.TypeName{
		tupleconv.TypeUnsigned,
		tupleconv.TypeDouble,
		tupleconv.TypeString,
	}

	converters := make([]tupleconv.Converter[string, any], 0)
	for _, typ := range fieldTypes {
		conv, _ := tupleconv.GetConverterByType[string](factory, typ)
		converters = append(converters, conv)
	}

	mapper := tupleconv.MakeMapper(converters)
	result, err := mapper.Map([]string{"1", "-2,2", "some_string"})
	fmt.Println(result, err)

	// Output:
	// [1 -2.2 some_string] <nil>
}

// ExampleStringToTTConvFactory_convertNullable demonstrates an example of converting
// a nullable type: an attempt to convert to null will be made before attempting to convert to the
// main type.
func ExampleStringToTTConvFactory_convertNullable() {
	factory := tupleconv.MakeStringToTTConvFactory().
		WithNullValue("2.5")

	converters, _ := tupleconv.MakeTypeToTTConverters[string](factory, []tupleconv.SpaceField{
		{Type: tupleconv.TypeDouble, IsNullable: true},
	})
	fmt.Println(converters[0].Convert("2.5"))

	// Output:
	// <nil> <nil>
}

type customFactory struct {
	tupleconv.StringToTTConvFactory
}

func (f *customFactory) MakeTypeToAnyMapper() tupleconv.Converter[string, any] {
	return tupleconv.MakeFuncConverter(func(s string) (any, error) {
		return s, nil
	})
}

// ExampleTTConvFactory_custom demonstrates how to customize the behavior of TTConvFactory
// by inheriting from it and overriding the necessary functions.
func ExampleTTConvFactory_custom() {
	facture := &customFactory{}
	spaceFmt := []tupleconv.SpaceField{{Type: tupleconv.TypeAny}}
	converters, _ := tupleconv.MakeTypeToTTConverters[string](facture, spaceFmt)

	res, err := converters[0].Convert("12")
	fmt.Println(res, err)

	// Output:
	// 12 <nil>
}

const workDir = "work_dir"
const server = "127.0.0.1:3014"

var dialer = tarantool.NetDialer{
	Address:  server,
	User:     "test",
	Password: "password",
}
var opts = tarantool.Opts{
	Timeout: 5 * time.Second,
}

func upTarantool() (func(), error) {
	inst, err := test_helpers.StartTarantool(test_helpers.StartOpts{
		Dialer:       dialer,
		InitScript:   "testdata/config.lua",
		Listen:       server,
		WorkDir:      workDir,
		WaitStart:    100 * time.Millisecond,
		ConnectRetry: 3,
		RetryTimeout: 500 * time.Millisecond,
	})
	if err != nil {
		test_helpers.StopTarantoolWithCleanup(inst)
		return nil, nil
	}

	cleanup := func() {
		test_helpers.StopTarantoolWithCleanup(inst)
	}
	return cleanup, nil
}

func makeTtEncoder() func(any) (string, error) {
	datetimeConverter := tupleconv.MakeDatetimeToStringConverter()
	return func(src any) (string, error) {
		switch src := src.(type) {
		case datetime.Datetime:
			return datetimeConverter.Convert(src)
		default:
			return fmt.Sprint(src), nil
		}
	}
}

// ExampleMap_insertMappedTuples demonstrates the combination of Mapper and go-tarantool
// functionality: firstly map string tuples to the tarantool types, then insert them to the
// target space.
func ExampleMap_insertMappedTuples() {
	cleanupTarantool, err := upTarantool()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanupTarantool()

	conn, _ := tarantool.Connect(context.Background(), dialer, opts)
	defer conn.Close()

	var spaceFmtResp [][]tupleconv.SpaceField
	req := tarantool.NewCallRequest("get_test_space_fmt")
	if err := conn.Do(req).GetTyped(&spaceFmtResp); err != nil {
		fmt.Printf("can't get target space fmt: %v\n", err)
		return
	}

	spaceFmt := spaceFmtResp[0]
	fmt.Println(spaceFmt[0:3])

	fac := tupleconv.MakeStringToTTConvFactory()
	converters, _ := tupleconv.MakeTypeToTTConverters[string](fac, spaceFmt)
	decoder := tupleconv.MakeMapper(converters).
		WithDefaultConverter(fac.GetStringConverter())

	dt1 := "2020-08-22T11:27:43.123456789-0200"
	dt2 := "1880-01-01T00:00:00-0000"
	uuid := "00000000-0000-0000-0000-000000000001"
	interval := "1,2,3,4,5,6,7,8,1"

	tuples := [][]string{
		{"1", "true", "12", "143.5", dt1, "", "str", "", "[1,2,3]", "190", ""},
		{"2", "f", "0", "-42", dt2, interval, "abacaba", "", "[]", uuid, "150"},

		// Extra fields.
		{"4", "1", "12", "143.5", dt1, "", "str", uuid, "[1,2,3]", "190", "", "extra", "blabla"},
	}

	for _, tuple := range tuples {
		mapped, err := decoder.Map(tuple)
		if err != nil {
			fmt.Println(err)
			return
		}
		insertReq := tarantool.NewInsertRequest("test_space").Tuple(mapped)
		_, err = conn.Do(insertReq).Get()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	selectReq := tarantool.NewSelectRequest("test_space")
	resp, err := conn.Do(selectReq).Get()
	if err != nil {
		fmt.Println(err)
		return
	}

	tuple0, _ := resp[0].([]any)
	encoder := tupleconv.MakeMapper[any, string]([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(tupleconv.MakeFuncConverter(makeTtEncoder()))

	encodedTuple0, _ := encoder.Map(tuple0)
	fmt.Println(encodedTuple0)

	// Output:
	// [{0 id unsigned false} {0 boolean boolean false} {0 number number false}]
	// [1 true 12 143.5 2020-08-22T11:27:43.123456789-0200 <nil> str <nil> [1 2 3] 190 <nil>]
}

// Example_ttEncoder demonstrates how to create an encoder, using Mapper with only a default
// Converter defined.
func Example_ttEncoder() {
	cleanupTarantool, err := upTarantool()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanupTarantool()

	converter := tupleconv.MakeFuncConverter(makeTtEncoder())
	tupleEncoder := tupleconv.MakeMapper([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(converter)

	conn, _ := tarantool.Connect(context.Background(), dialer, opts)
	defer conn.Close()

	req := tarantool.NewSelectRequest("finances")

	var tuples [][]any
	if err := conn.Do(req).GetTyped(&tuples); err != nil {
		fmt.Printf("can't select tuples: %v\n", err)
		return
	}

	for _, tuple := range tuples {
		encoded, err := tupleEncoder.Map(tuple)
		if err != nil {
			fmt.Printf("can't encode tuple: %v\n", err)
			return
		}
		fmt.Println(encoded)
	}

	// Output:
	// [1 14.15 2023-08-30T12:13:00+0000]
	// [2 193000 2023-08-31T00:00:00 Europe/Paris]
	// [3 -111111 2023-09-02T00:01:00+0400]
}
