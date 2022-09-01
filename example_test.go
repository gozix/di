// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di_test

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gozix/di"
)

// This is example demonstrates Container.Resolve usage with aliases.
func Example_interfaces() {
	var builder, err = di.NewBuilder(
		// provide constructor
		di.Provide(NewBarController, di.As(new(Controller))),
		di.Provide(NewBazController, di.As(new(Controller))),
	)

	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var container di.Container
	if container, err = builder.Build(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var controllers []Controller
	if err = container.Resolve(&controllers); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("resolved", len(controllers), "controllers")

	// Output:
	// resolved 2 controllers
}

// This is example demonstrates Container.Has usage with modifiers.
func Example_tags() {
	var builder, err = di.NewBuilder(
		// provide constructor
		di.Provide(NewBarController, di.Tags{{
			Name: "controller",
		}}),
	)

	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var container di.Container
	if container, err = builder.Build(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if container.Has((*BarController)(nil), di.WithTags("controller")) {
		fmt.Println("container has *BarController with tag controller")
	}

	var c *BarController
	if err = container.Resolve(&c, di.WithTags("controller")); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if c != nil {
		fmt.Println("container resolved *BarController with tag controller")
	}

	err = container.Call(func(c *BarController) error {
		if c == nil {
			return errors.New("container call has nil *BarController")
		}

		fmt.Println("container call *BarController with tag controller")

		return nil
	}, di.Constraint(0, di.WithTags("controller")))

	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// Output:
	// container has *BarController with tag controller
	// container resolved *BarController with tag controller
	// container call *BarController with tag controller
}

// This is example demonstrates optional field injection.
func Example_optionalFieldInjection() {
	type Foo struct {
		Bar *BarController
		Baz *BazController
	}

	var builder, err = di.NewBuilder(
		// provide autowired type with optional field Baz
		di.Autowire((*Foo)(nil), di.Constraint("Baz", di.Optional(true))),
		// provide constructor
		di.Provide(NewBarController),
	)

	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var container di.Container
	if container, err = builder.Build(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	err = container.Call(func(foo *Foo) error {
		if foo == nil {
			return errors.New("foo is nil")
		}

		if foo.Bar == nil {
			return errors.New("foo.Bar is nil")
		}

		if foo.Baz != nil {
			return errors.New("foo.Baz is not nil")
		}

		return nil
	})

	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Looks like something is fine")

	// Output:
	// Looks like something is fine
}

// This is example demonstrates how use dependency container for simple http application.
func Example_httpServer() {
	var builder, err = di.NewBuilder(
		di.BuilderOptions(
			di.Autowire((*BarController)(nil), di.As((*Controller)(nil))), // provide autowired type
			di.Autowire((*BazController)(nil), di.As((*Controller)(nil))), // provide autowired type
		),
		di.BuilderOptions(
			di.Provide(NewServerMux), // provide constructor
			di.Provide(NewServer),    // provide constructor
		),
	)

	if err != nil {
		fmt.Println()
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// build container
	var container di.Container
	if container, err = builder.Build(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// Call function with resolved dependencies
	_ = container.Call(func(srv *http.Server) {
		if srv == nil {
			_, _ = fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	})

	fmt.Println("Listening...")

	// Output:
	// Listening...
}
