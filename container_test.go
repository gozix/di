// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gozix/di"
	"github.com/gozix/di/internal/cycle"

	"github.com/stretchr/testify/require"
)

func NewContainer() (di.Container, error) {
	var builder, err = di.NewBuilder(
		di.Provide(
			NewServerMux,
			di.Constraint(
				([]Controller)(nil),
				di.Optional(false),
				di.WithTags("controller"),
				di.WithoutTags("cycled", "flaky"),
			),
		),
		di.Autowire((*BarController)(nil), di.As(new(Controller)), di.Unshared(), di.Tags{{
			Name: "controller",
		}, {
			Name: "bar",
		}}),
		di.Add(NewBazController(), di.As(new(Controller)), di.Tags{{
			Name: "controller",
		}, {
			Name: "baz",
		}}),
		di.Provide(NewCycledController, di.As(new(Controller)), di.Unshared(), di.Tags{{
			Name: "controller",
		}, {
			Name: "cycled",
		}}),
		di.Provide(NewFlakyController, di.As(new(Controller)), di.Tags{{
			Name: "controller",
		}, {
			Name: "flaky",
		}}),
	)

	if err != nil {
		return nil, err
	}

	return builder.Build()
}

func TestContainer(t *testing.T) {
	type TestCase struct {
		Name string
		Run  func(t *testing.T, ctn di.Container)
	}

	var testCases = []TestCase{{
		Name: "Call with invalid argument",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(nil)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrorMustBeFunction)
			require.ErrorContains(t, err, "container_test.go:68")
		},
	}, {
		Name: "Call with unregistered type",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(srv *http.Server) {
				_ = srv
			})

			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrDoesNotExist)
			require.ErrorContains(t, err, "container_test.go:76")
		},
	}, {
		Name: "Call without dependencies",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(bar *BarController) {
				require.NotNil(t, bar)
			})

			require.NoError(t, err)
		},
	}, {
		Name: "Call with multiple dependencies",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(controllers []Controller) {
				require.NotNil(t, controllers)
				require.Len(t, controllers, 2)
			}, di.Constraint(
				0,
				di.Optional(true),
				di.WithTags("controller"),
				di.WithoutTags("cycled", "flaky"),
			))

			require.NoError(t, err)
		},
	}, {
		Name: "Call with constraint modifiers",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(controllers []Controller) {
				require.Len(t, controllers, 1)
			}, di.Constraint(([]Controller)(nil), di.WithTags("bar")))

			require.NoError(t, err)
		},
	}, {
		Name: "Call with constraint optional and modifiers",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(controllers []Controller) {
				require.Len(t, controllers, 0)
			}, di.Constraint(([]Controller)(nil), di.Optional(true), di.WithTags("not exist")))

			require.NoError(t, err)
		},
	}, {
		Name: "Call with same type arguments",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(bar1, bar2 *BarController) {
				require.NotNil(t, bar1)
				require.NotNil(t, bar2)
			})

			require.NoError(t, err)
		},
	}, {
		Name: "Call with variadic args",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(
				func(controllers ...Controller) {
					require.Len(t, controllers, 2)
				},
				di.Constraint(0, di.WithTags("controller"), di.WithoutTags("cycled", "flaky")),
			)

			require.NoError(t, err)
		},
	}, {
		Name: "Call with Container argument",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func(container di.Container) {
				require.NotNil(t, container)
				require.Same(t, ctn, container)
			})

			require.NoError(t, err)
		},
	}, {
		Name: "Call func with error",
		Run: func(t *testing.T, ctn di.Container) {
			var eErr = errors.New("expected")
			var aErr = ctn.Call(func(baz *BazController) error {
				require.NotNil(t, baz)
				return eErr
			})

			require.ErrorIs(t, aErr, eErr)
			require.ErrorContains(t, aErr, "container_test.go:162")
		},
	}, {
		Name: "Has dependency exist",
		Run: func(t *testing.T, ctn di.Container) {
			require.True(t, ctn.Has((*BarController)(nil)))
		},
	}, {
		Name: "Has dependency absent",
		Run: func(t *testing.T, ctn di.Container) {
			require.False(t, ctn.Has((*http.Server)(nil)))
		},
	}, {
		Name: "Resolve invalid error",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Resolve(nil)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrMustBeSliceOrPointer)
			require.ErrorContains(t, err, "container_test.go:183")
		},
	}, {
		Name: "Resolve multiple error",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				foo Controller
				err = ctn.Resolve(&foo)
			)

			require.Nil(t, foo)
			require.Error(t, err)
			require.ErrorIs(t, err, di.ErrMultipleDefinitions)
			require.ErrorContains(t, err, "container_test.go:193")
		},
	}, {
		Name: "Resolve flaky dependency",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				flaky *FlakyController
				err   = ctn.Resolve(&flaky)
			)

			require.Nil(t, flaky)
			require.Error(t, err)
			require.ErrorContains(t, err, "container_test.go:206")
		},
	}, {
		Name: "Resolve cycled",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				srv *CycledController
				err = ctn.Resolve(&srv)
			)

			require.Nil(t, srv)
			require.ErrorIs(t, err, cycle.ErrCycleDetected)
			require.ErrorContains(t, err, "container_test.go:218")
		},
	}, {
		Name: "Resolved group with flaky item",
		Run: func(t *testing.T, ctn di.Container) {
			var err = ctn.Call(func([]Controller) {
				require.NoError(t, errors.New("don't be coled"))
			}, di.Constraint(([]Controller)(nil)))

			require.Error(t, err)
			require.ErrorContains(t, err, "container_test.go:228")
		},
	}, {
		Name: "Resolve without dependencies",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				bar *BarController
				err = ctn.Resolve(&bar)
			)

			require.NoError(t, err)
			require.NotNil(t, bar)
		},
	}, {
		Name: "Resolve with dependencies",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				srv *http.ServeMux
				err = ctn.Resolve(&srv)
			)

			require.NoError(t, err)
			require.NotNil(t, srv)
		},
	}, {
		Name: "Resolve shared instance",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				srv1, srv2 *http.ServeMux
				err        error
			)

			err = ctn.Resolve(&srv1)
			require.NoError(t, err)
			require.NotNil(t, srv1)

			err = ctn.Resolve(&srv2)
			require.NoError(t, err)
			require.NotNil(t, srv2)

			require.Same(t, srv1, srv2)
		},
	}, {
		Name: "Resolve unshared instance",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				bar1, bar2 *BarController
				err        error
			)

			err = ctn.Resolve(&bar1)
			require.NoError(t, err)
			require.NotNil(t, bar1)

			err = ctn.Resolve(&bar2)
			require.NoError(t, err)
			require.NotNil(t, bar2)

			require.NotSame(t, bar1, bar2)
		},
	}, {
		Name: "Resolve single instance by tag",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				bar Controller
				err = ctn.Resolve(&bar, di.WithTags("bar"))
			)

			require.NoError(t, err)
			require.NotNil(t, bar)
			require.IsType(t, (*BarController)(nil), bar)
		},
	}, {
		Name: "Resolve multiple instances by tag",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				controllers []Controller
				err         = ctn.Resolve(&controllers, di.WithTags("controller"), di.WithoutTags("cycled", "flaky"))
			)

			require.NoError(t, err)
			require.NotNil(t, controllers)
			require.Len(t, controllers, 2)
		},
	}, {
		Name: "Resolve container",
		Run: func(t *testing.T, ctn di.Container) {
			var (
				container di.Container
				err       = ctn.Resolve(&container)
			)

			require.NoError(t, err)
			require.NotNil(t, container)
			require.Same(t, ctn, container)
		},
	}}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase#%d: %s", i+1, testCase.Name), func(t *testing.T) {
			var c, err = NewContainer()
			require.NoError(t, err)

			testCase.Run(t, c)

			err = c.Close()
			require.NoError(t, err)
		})
	}
}
