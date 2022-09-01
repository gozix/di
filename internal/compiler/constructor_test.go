// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package compiler_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gozix/di/internal/compiler"

	"github.com/stretchr/testify/require"
)

func TestConstructor(t *testing.T) {
	type (
		Result struct {
			Closer compiler.Closer
			Error  error
			Value  reflect.Value
		}

		TestCase = struct {
			Constructor  any
			Dependencies []*compiler.Dependency
			Result       *Result
			Error        error
			Type         reflect.Type
		}
	)

	var closer = func() error {
		return nil
	}

	var testCases = []TestCase{{
		Constructor: []string{},
		Error:       compiler.ErrInvalidConstructor,
	}, {
		Constructor: func() {},
		Error:       compiler.ErrInvalidConstructor,
	}, {
		Constructor: func() (int, int, int, int) {
			return 0, 0, 0, 0
		},
		Error: compiler.ErrInvalidConstructor,
	}, {
		Constructor: func(a, b int) int {
			return a + b
		},
		Dependencies: []*compiler.Dependency{{
			Name:  "int",
			Index: 0,
			Type:  reflect.TypeOf(1),
			Value: reflect.ValueOf(1),
		}, {
			Name:  "int",
			Index: 1,
			Type:  reflect.TypeOf(1),
			Value: reflect.ValueOf(1),
		}},
		Result: &Result{
			Value: reflect.ValueOf(2),
		},
		Type: reflect.TypeOf(1),
	}, {
		Constructor: func() (int, error) {
			return 0, errors.New("oops")
		},
		Dependencies: []*compiler.Dependency{},
		Result: &Result{
			Error: errors.New("oops"),
			Value: reflect.ValueOf(0),
		},
		Type: reflect.TypeOf(0),
	}, {
		Constructor: func() (int, func() error) {
			return 0, closer
		},
		Dependencies: []*compiler.Dependency{},
		Result: &Result{
			Closer: closer,
			Value:  reflect.ValueOf(0),
		},
		Type: reflect.TypeOf(0),
	}, {
		Constructor: func() (int, func() error, error) {
			return 0, closer, nil
		},
		Dependencies: []*compiler.Dependency{},
		Result: &Result{
			Value:  reflect.ValueOf(0),
			Closer: closer,
		},
		Type: reflect.TypeOf(0),
	}, {
		Constructor: func() (int, func() error, error) {
			return 0, closer, errors.New("oops")
		},
		Dependencies: []*compiler.Dependency{},
		Result: &Result{
			Closer: closer,
			Error:  errors.New("oops"),
			Value:  reflect.ValueOf(0),
		},
		Type: reflect.TypeOf(0),
	}, {
		Constructor: func(values ...int) int {
			var val = 0
			for _, v := range values {
				val += v
			}

			return val
		},
		Dependencies: []*compiler.Dependency{{
			Name:  "",
			Index: 0,
			Type:  reflect.TypeOf([]int{}),
			Value: reflect.ValueOf([]int{1, 2}),
		}},
		Result: &Result{
			Value: reflect.ValueOf(3),
		},
		Type: reflect.TypeOf(0),
	}}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			var cmp, err = compiler.NewConstructor(testCase.Constructor)
			if err != nil && errors.Is(err, testCase.Error) {
				return
			}

			require.NoError(t, err)

			var typ = cmp.Type()
			require.Equal(t, testCase.Type.String(), typ.String())

			var deps = cmp.Dependencies()
			require.Conditionf(t, func() (success bool) {
				for j := range deps {
					if j >= len(testCase.Dependencies) {
						return false
					}

					var equal = deps[j].Name == testCase.Dependencies[j].Name &&
						deps[j].Index == testCase.Dependencies[j].Index &&
						deps[j].Type.String() == testCase.Dependencies[j].Type.String()

					if !equal {
						return false
					}

					if !testCase.Dependencies[j].Value.IsValid() {
						return false
					}
				}

				return true
			}, "Dependencies not equal")

			var v, c, e = cmp.Create(testCase.Dependencies...)
			require.Equal(t, testCase.Result.Error, e)
			require.Equal(t, testCase.Result.Value.Interface(), v.Interface())
			require.Equal(t, reflect.ValueOf(testCase.Result.Closer).Pointer(), reflect.ValueOf(c).Pointer())
		})
	}
}
