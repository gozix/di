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

type (
	Bar struct {
		private1 int
	}

	Baz struct {
		Public1  int
		Public2  int
		private1 int
	}
)

func TestType(t *testing.T) {
	type (
		TestCase = struct {
			Dependencies []*compiler.Dependency
			Error        error
			Source       any
			Type         reflect.Type
		}
	)

	var testCases = []TestCase{{
		Source: (*Bar)(nil),
		Type:   reflect.TypeOf((*Bar)(nil)),
	}, {
		Dependencies: []*compiler.Dependency{{
			Name:  "Public1",
			Index: 0,
			Type:  reflect.TypeOf(0),
			Value: reflect.ValueOf(0),
		}, {
			Name:  "Public2",
			Index: 1,
			Type:  reflect.TypeOf(0),
			Value: reflect.ValueOf(0),
		}},
		Source: (*Baz)(nil),
		Type:   reflect.TypeOf((*Baz)(nil)),
	}, {
		Source: 0,
		Type:   reflect.TypeOf(0),
	}, {
		Source: nil,
		Error:  compiler.ErrInvalidType,
	}}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			var cmp, err = compiler.NewType(testCase.Source)
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
			require.Equal(t, testCase.Type.String(), v.Type().String())
			require.Nil(t, c)
			require.Nil(t, e)
		})
	}
}
