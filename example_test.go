package tupleconv_test

import (
	"errors"
	"fmt"
	"github.com/tarantool/go-tarantool"
	"github.com/tarantool/go-tarantool/datetime"
	"github.com/tarantool/go-tarantool/test_helpers"
	"github.com/tarantool/go-tupleconv"
	"strconv"
	"strings"
	"time"

	_ "github.com/tarantool/go-tarantool/uuid"
)

type filterIntConverter struct {
	toFilter string
}

func (c filterIntConverter) Convert(src string) (int64, error) {
	src = strings.ReplaceAll(src, c.toFilter, "")
	return strconv.ParseInt(src, 10, 64)
}

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

func ExampleMapper_singleMapper() {
	// Single mapper example.
	toStringMapper := tupleconv.MakeMapper([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(tupleconv.MakeFuncConverter(
			func(s any) (string, error) {
				return fmt.Sprintln(s), nil
			}),
		)
	res, err := toStringMapper.Map([]any{1, 2.5, nil})
	fmt.Println(res, err)

	// Output:
	// [1
	//  2.5
	//  <nil>
	//] <nil>
}

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

func upTarantool() (func(), error) {
	inst, err := test_helpers.StartTarantool(test_helpers.StartOpts{
		InitScript:   "testdata/config.lua",
		Listen:       server,
		WorkDir:      workDir,
		User:         "test",
		Pass:         "password",
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

func ExampleMap_insertMappedTuples() {
	cleanupTarantool, err := upTarantool()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanupTarantool()

	conn, _ := tarantool.Connect(server, tarantool.Opts{
		User: "test",
		Pass: "password",
	})
	var spaceFmtResp [][]tupleconv.SpaceField
	_ = conn.CallTyped("get_test_space_fmt", []any{}, &spaceFmtResp)
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
		resp, err := conn.Do(insertReq).Get()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("insert response code =", resp.Code)
	}

	selectReq := tarantool.NewSelectRequest("test_space")
	resp, err := conn.Do(selectReq).Get()
	if err != nil {
		fmt.Println(err)
		return
	}

	tuple0, _ := resp.Data[0].([]any)
	encoder := tupleconv.MakeMapper[any, string]([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(tupleconv.MakeFuncConverter(func(s any) (string, error) {
			asDatetime, isDatetime := s.(datetime.Datetime)
			if isDatetime {
				return fmt.Sprintln(asDatetime.ToTime()), nil
			} else {
				return fmt.Sprintln(s), nil
			}
		}))

	encodedTuple0, _ := encoder.Map(tuple0)
	fmt.Println(encodedTuple0)

	// Output:
	// [{0 id unsigned false} {0 boolean boolean false} {0 number number false}]
	// insert response code = 0
	// insert response code = 0
	// insert response code = 0
	// [1
	//  true
	//  12
	//  143.5
	//  2020-08-22 11:27:43.123456789 -0200 -0200
	//  <nil>
	//  str
	//  <nil>
	//  [1 2 3]
	//  190
	//  <nil>
	//]
}
