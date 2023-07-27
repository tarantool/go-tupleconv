package tupleconv_test

import (
	"errors"
	"fmt"
	"github.com/tarantool/go-tupleconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapper_singleMapper(t *testing.T) {
	stringer := tupleconv.MakeFuncConverter(func(t any) (string, error) {
		return fmt.Sprintln(t), nil
	})

	encoder := tupleconv.MakeMapper([]tupleconv.Converter[any, string]{}).
		WithDefaultConverter(stringer)

	cases := []struct {
		name     string
		tuple    []any
		expected []string
	}{
		{
			name:     "empty",
			tuple:    []any{},
			expected: []string{},
		},
		{
			name: "different types",
			tuple: []any{
				"a",
				1,
				nil,
				map[string]string{
					"1": "2",
					"3": "4",
				},
			},
			expected: []string{
				"a\n",
				"1\n",
				"<nil>\n",
				"map[1:2 3:4]\n",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := encoder.Map(tc.tuple)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMapper_basicMapper(t *testing.T) {
	idealMapper := tupleconv.MakeFuncConverter(func(t string) (any, error) {
		return 42, nil
	})

	someError := errors.New("some error")
	nonIdealMapper := tupleconv.MakeFuncConverter(func(t string) (any, error) {
		if t == "bad" {
			return "", someError
		}
		return t, nil
	})

	decoder := tupleconv.MakeMapper([]tupleconv.Converter[string, any]{idealMapper, nonIdealMapper})

	cases := []struct {
		name          string
		tuple         []string
		expectedTuple []any
		wantErr       bool
	}{
		{
			name:          "all is ok",
			tuple:         []string{"1", "2"},
			expectedTuple: []any{42, "2"},
		},
		{
			name:    "too big tuple length",
			tuple:   []string{"1", "2", "3", "4"},
			wantErr: true,
		},
		{
			name:    "decoding error",
			tuple:   []string{"1", "bad"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualTuple, err := decoder.Map(tc.tuple)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTuple, actualTuple)
			}
		})
	}
}

func TestMapper_partialMapper(t *testing.T) {
	mapper := tupleconv.MakeMapper([]tupleconv.Converter[string, any]{
		tupleconv.MakeFuncConverter(func(s string) (any, error) {
			return 42, nil
		}),
		tupleconv.MakeFuncConverter(func(s string) (any, error) {
			if s == "bad" {
				return "", fmt.Errorf("error")
			}
			return s, nil
		}),
	}).WithDefaultConverter(tupleconv.MakeFuncConverter(func(s string) (any, error) {
		if s == "null" {
			return nil, nil
		}
		return s, nil
	}))

	cases := []struct {
		name          string
		tuple         []string
		expectedTuple []any
		wantErr       bool
	}{
		{
			name:          "all is ok",
			tuple:         []string{"1", "2"},
			expectedTuple: []any{42, "2"},
		},
		{
			name:    "decoding error",
			tuple:   []string{"1", "bad"},
			wantErr: true,
		},
		{
			name:          "long tuple",
			tuple:         []string{"1", "2", "3", "4", "null"},
			expectedTuple: []any{42, "2", "3", "4", nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualTuple, err := mapper.Map(tc.tuple)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTuple, actualTuple)
			}
		})
	}
}
