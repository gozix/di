// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di_test

import (
	"testing"

	"github.com/gozix/di"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	type testCase struct {
		Name   string
		Runner func(t *testing.T, builder di.Builder)
	}

	var testCases = []testCase{{
		Name: "Builder -> Add",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Add(BarController{})
			require.NoError(t, err)
		},
	}, {
		Name: "Builder -> Add with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Add(BarController{}, di.As(nil))
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrIsNil)
			require.ErrorContains(t, err, "builder_test.go:30")
		},
	}, {
		Name: "Builder -> Add with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Add(nil)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidValue)
			require.ErrorContains(t, err, "builder_test.go:38")
		},
	}, {
		Name: "Builder -> Apply -> Add with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.Add(nil),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidValue)
			require.ErrorContains(t, err, "builder_test.go:47")
		},
	}, {
		Name: "Builder -> Apply -> BuilderOptions -> Add with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.BuilderOptions(
					di.Add(nil),
				),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidValue)
			require.ErrorContains(t, err, "builder_test.go:58")
		},
	}, {
		Name: "Builder -> Autowire",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Autowire((*BarController)(nil))
			require.NoError(t, err)
		},
	}, {
		Name: "Builder -> Autowire with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Autowire(nil)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidType)
			require.ErrorContains(t, err, "builder_test.go:74")
		},
	}, {
		Name: "Builder -> Apply -> Autowire with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.Autowire(nil),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidType)
			require.ErrorContains(t, err, "builder_test.go:83")
		},
	}, {
		Name: "Builder -> Apply -> BuilderOptions -> Autowire with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.BuilderOptions(
					di.Autowire(nil),
				),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidType)
			require.ErrorContains(t, err, "builder_test.go:94")
		},
	}, {
		Name: "Builder -> Provide",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Provide(NewBarController)
			require.NoError(t, err)
		},
	}, {
		Name: "Builder -> Provide with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Provide(nil)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidConstructor)
			require.ErrorContains(t, err, "builder_test.go:110")
		},
	}, {
		Name: "Builder -> Apply -> Provide with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.Provide(nil),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidConstructor)
			require.ErrorContains(t, err, "builder_test.go:119")
		},
	}, {
		Name: "Builder -> Apply -> BuilderOptions -> Provide with error",
		Runner: func(t *testing.T, builder di.Builder) {
			var err = builder.Apply(
				di.BuilderOptions(
					di.Provide(nil),
				),
			)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrInvalidConstructor)
			require.ErrorContains(t, err, "builder_test.go:130")
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var builder, err = di.NewBuilder()
			require.NoError(t, err)

			tc.Runner(t, builder)

			var container di.Container
			container, err = builder.Build()
			require.NoError(t, err)
			require.NotNil(t, container)
		})
	}
}
