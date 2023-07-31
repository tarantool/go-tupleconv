# Tarantool tuples converter in Go

[![Go Reference][godoc-badge]][godoc-url]
[![Actions Status][actions-badge]][actions-url]
[![Code Coverage][coverage-badge]][coverage-url]

## Table of contents
* [Documentation](#documentation)
  * [Converter](#converter)
  * [Mapper](#mapper)
  * [Mappers to tarantool types](#mappers-to-tarantool-types)
    * [Example](#example)
    * [String to nullable](#string-to-nullable)
    * [String to any/scalar](#string-to-anyscalar)
    * [Customization](#customization)
## Documentation

### Converter
`Converter[S,T]` converts objects of type `S` into objects of type `T`. Converters
are basic entities on which mappers are based.  
Implementations of some converters are available, for example, converters
from strings to golang types.   
Usage example:
```golang
// Basic converter.
strToBoolConv := tupleconv.MakeStringToBoolConverter()
result, err := strToBoolConv.Convert("true") // true <nil>

// Function based converter.
funcConv := tupleconv.MakeFuncConverter(func(s string) (string, error) {
    return s + " world!", nil
})
result, err = funcConv.Convert("hello") // hello world! <nil>
```
**Note 1**: You can use the provided converters.

**Note 2**: You can create your own converters based on the functions 
with `tupleconv.MakeFuncConverter`.

**Note 3**: You can create your own converters, implementing
`Converter[S,T]` interface.

### Mapper
`Mapper` is an object that converts tuples. It is built using a list of 
converters.  
Usage example:
```golang
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
result, err := mapper.Map([]string{"a", "4"}) // []any{"a1", 5} <nil>
result, err = mapper.Map([]string{"0"}) // []any{"01"} <nil>
```
```golang
// Single mapper example.
toStringMapper := tupleconv.MakeMapper([]tupleconv.Converter[any, string]{}).
        WithDefaultConverter(tupleconv.MakeFuncConverter(
            func(s any) (string, error) {
                return fmt.Sprintln(s), nil
        }),
)
res, err := toStringMapper.Map([]any{1, 2.5, nil}) // ["1\n", "2.5\n", "<nil>\n"] <nil>
```
**Note 1**: To create a mapper, an array of converters is needed, each
of which transforms a certain type S into type T.   

**Note 2**: To perform tuple mapping, you can use the function 
`Map`, which will return control to the calling code upon the first error.  

**Note 3**: You can set a default converter that will be applied if the tuple length exceeds
the size of the primary converters list.   
For example, if you only set a default converter, `Map` will work like the `map` function in
functional programming languages.

**Note 4**: If tuple length is less than converters list length, then only corresponding converters
will be applied.

### Mappers to tarantool types

#### Example
For building an array of converters, especially when it comes to conversions to 
tarantool types, there is a built-in solution.  
Let's consider an example:
```golang
factory := tupleconv.MakeStringToTTConvFactory().
                WithDecimalSeparators(",.")

spaceFmt := []tupleconv.SpaceField{
    {Type: tupleconv.TypeUnsigned},
    {Type: tupleconv.TypeDouble, IsNullable: true},
    {Type: tupleconv.TypeString},
}

converters, _ := tupleconv.MakeTypeToTTConverters[string](factory, spaceFmt)
mapper := tupleconv.MakeMapper(converters)
result, err := mapper.Map([]string{"1", "-2,2", "some_string"}) // [1, -2.2, "some_string"] <nil>
```
**Note 1**: To build an array of converters, the space format and a 
certain object implementing `TTConvFactory` are used. Function 
`MakeTypeToTTConverters` takes these entities and gives the converters list.

**Note 2**: `TTConvFactory[Type]` is capable of building a 
converter from `Type` to each tarantool type.

**Note 3**: There is a basic factory available called 
`StringToTTConvFactory`, which is used for conversions from strings to 
tarantool types.   

**Note 4**: `StringToTTConvFactory` can be configured with options like
`WithDecimalSeparators`.

#### String to nullable
When converting nullable types with `StringToTTConvFactory`, first, an attempt
is made to convert to null.

For example, empty string is interpreted like `null` with default options.
If a field has a `string` type and is `nullable`, then an empty string will be
converted to null during the conversion process, rather than being
converted to empty string.


#### String to any/scalar
When converting to `any`/`scalar` with `StringToTTConvFactory`, by default, 
an attempt will be made to convert them to the following types, 
in the following order:
- `number`
- `decimal`
- `boolean`
- `datetime`
- `uuid`
- `interval`
- `string`

#### Customization
`TTConvFactory[Type]` is an interface that can build a mapper from 
`Type` to each tarantool type.   
To customize the behavior for specific types, one can 
inherit from the existing factory and override the necessary methods.  
For example, let's make the standard factory for conversion from strings to 
tarantool types always convert `any` type to a string:
```golang
type customFactory struct {
    tupleconv.StringToTTConvFactory
}

func (f *customFactory) MakeTypeToAnyMapper() tupleconv.Converter[string, any] {
    return tupleconv.MakeFuncConverter(func(s string) (any, error) {
        return s, nil
    })
}

func example() {
    factory := &customFactory{}
    spaceFmt := []tupleconv.SpaceField{{Type: "any"}}
    converters, _ := tupleconv.MakeTypeToTTConverters[string](factory, spaceFmt)

    res, err := converters[0].Convert("12") // "12" <nil>
}
```

[godoc-badge]: https://pkg.go.dev/badge/github.com/tarantool/go-tupleconv.svg
[godoc-url]: https://pkg.go.dev/github.com/tarantool/go-tupleconv
[actions-badge]: https://github.com/tarantool/go-tupleconv/actions/workflows/test.yml/badge.svg
[actions-url]: https://github.com/tarantool/go-tupleconv/actions/workflows/test.yml
[coverage-badge]: https://coveralls.io/repos/github/tarantool/go-tupleconv/badge.svg?branch=master
[coverage-url]: https://coveralls.io/github/tarantool/go-tupleconv?branch=master