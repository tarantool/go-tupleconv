package tupleconv

import (
	"fmt"
)

// Mapper performs tuple mapping.
type Mapper[S any, T any] struct {
	converters       []Converter[S, T]
	defaultConverter Converter[S, T]
}

// MakeMapper creates Mapper.
func MakeMapper[S any, T any](converters []Converter[S, T]) Mapper[S, T] {
	return Mapper[S, T]{converters: converters}
}

// WithDefaultConverter sets defaultConverter. A nil converter unsets it.
func (mapper Mapper[S, T]) WithDefaultConverter(converter Converter[S, T]) Mapper[S, T] {
	mapper.defaultConverter = converter
	return mapper
}

// validateTuple validates tuple in accordance with the Mapper properties.
func (mapper Mapper[S, T]) validateTuple(tuple []S) error {
	if len(tuple) > len(mapper.converters) && mapper.defaultConverter == nil {
		return fmt.Errorf("tuple length should be less or equal converters list length, " +
			"when default converter is not used")
	}
	return nil
}

// Map maps tuple until the first error.
func (mapper Mapper[S, T]) Map(tuple []S) ([]T, error) {
	if err := mapper.validateTuple(tuple); err != nil {
		return nil, err
	}
	var err error
	result := make([]T, len(tuple))
	for i, field := range tuple {
		converter := mapper.defaultConverter
		if i < len(mapper.converters) {
			converter = mapper.converters[i]
		}
		if converter == nil {
			return nil, fmt.Errorf("converter %d is nil", i)
		}
		result[i], err = converter.Convert(field)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
