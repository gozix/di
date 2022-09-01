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

func TestValue(t *testing.T) {
	type (
		Result struct {
			Closer compiler.Closer
			Error  error
			Value  reflect.Value
		}

		TestCase struct {
			Error  error
			Object *Result
			Value  any
			Type   reflect.Type
		}
	)

	var testCases = []TestCase{{
		Value: struct{}{},
		Object: &Result{
			Value: reflect.ValueOf(struct{}{}),
		},
		Type: reflect.TypeOf(struct{}{}),
	}, {
		Value: nil,
		Error: compiler.ErrInvalidValue,
	}}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase#%d", i+1), func(t *testing.T) {
			var cmp, err = compiler.NewValue(testCase.Value)
			if err != nil && errors.Is(err, testCase.Error) {
				return
			}

			require.NoError(t, err)

			var typ = cmp.Type()
			require.Equal(t, testCase.Type, typ)

			var deps = cmp.Dependencies()
			require.Equal(t, ([]*compiler.Dependency)(nil), deps)

			var v, c, e = cmp.Create(deps...)
			require.Equal(t, testCase.Object.Value.Interface(), v.Interface())
			require.Nil(t, c)
			require.Nil(t, e)
		})
	}
}
