// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"errors"
	"reflect"

	"github.com/gozix/di/internal/compiler"
)

type (
	// Builder must be used to create a Container. The Builder should be created with NewBuilder.
	// Then you can provide any definition with various allowed methods
	// and finally build the Container with the Build method.
	Builder interface {
		// Add provides value as is.
		//
		// The value argument must contain any value what you want, fre example:
		//   - new(http.Server)
		//   - *http.Server{}
		//   - etc.
		// The options argument may be one of:
		//   - di.Tags{}
		//   - di.As()
		Add(value Value, options ...AddOption) error

		// Apply applies options to Builder.
		// The options argument may be one of:
		//   - di.BuilderOptions()
		//   - di.Add()
		//   - di.Autowire()
		//   - di.Provide()
		Apply(options ...BuilderOption) error

		// Autowire providers autowired type.
		//
		// The target argument must contain the wanted type, for example:
		//   - (*http.Server)(nil)
		//   - (*io.Writer)(nil)
		//   - new(io.Writer)
		//   - etc.
		// The options argument may be one of:
		//   - di.As()
		//   - di.Constraint()
		//   - di.Tags{}
		//   - di.Unshared()
		Autowire(target Type, options ...ProvideOption) error

		// Provide provides any constructor.
		//
		// The constructor argument must be a function with one of the following signatures:
		//   - func New(constraints ...any) (value any)
		//   - func New(constraints ...any) (value any, err error)
		//   - func New(constraints ...any) (value any, closer func(){})
		//   - func New(constraints ...any) (value any, closer func(){}, err error)
		// The options argument may be one of:
		//   - di.As()
		//   - di.Constraint()
		//   - di.Tags{}
		//   - di.Unshared()
		Provide(constructor Constructor, options ...ProvideOption) error

		// Build is container build method.
		Build() (Container, error)

		// Definitions are build snapshot of definitions.
		Definitions() []Definition
	}

	// Constructor is any constructor.
	Constructor any

	// Container represents a dependency injection container.
	// To create a container, you should use a builder.
	Container interface {
		// Call calls the function with resolved arguments.
		//
		// The fn argument must contain any function. If the function contains error in the last return type,
		// then Call will return that value as own return type value.
		// The options argument may be one of:
		//   - di.Constraint()
		Call(fn Function, options ...ConstraintOption) (err error)

		// Close runs closers in reverse order that has been created.
		//
		// Any close function can return any error that stop the calling loop for all rest closers. Any close function
		// can return any error that stop the calling loop for all rest closers. That error will return in function
		// return type.
		Close() error

		// Has checks that type exists in container, if not it return false.
		//
		// The value argument must contain the wanted type, for example:
		//   - (*http.Server)(nil)
		//   - (*io.Writer)(nil)
		//   - new(io.Writer)
		//   - etc.
		// The modifiers argument may be one of:
		//   - di.WithTags()
		Has(value Type, modifiers ...Modifier) (exist bool)

		// Resolve resolves type and fills target pointer.
		//
		// The target argument must contain reference to wanted variable.
		// The modifiers argument may be one of:
		//   - di.WithTags()
		Resolve(target Value, modifiers ...Modifier) (err error)
	}

	// Definition represent container definition.
	Definition interface {
		// ID is definition unique identificator getter.
		ID() int

		// Dependencies is definition type dependencies getter.
		Dependencies() []Dependency

		// Tags is definition tags getter.
		Tags() Tags

		// Type is definition type getter.
		Type() reflect.Type

		// Unshared is definition unshared getter.
		Unshared() bool
	}

	// Dependency represent definition dependency.
	Dependency struct {
		// Type is type of dependency.
		Type reflect.Type

		// Optional is optional flag.
		Optional bool

		// Definitions are list of matched definitions.
		Definitions []Definition
	}

	// Function is any function.
	Function any

	// Type is any type.
	Type any

	// Value is any value.
	Value any
)

var (
	// ErrNotPointerToInterface is error triggered when provided alias not pointer to interface.
	ErrNotPointerToInterface = errors.New("not pointer to interface")

	// ErrIsNil is error triggered when provided nil alias.
	ErrIsNil = errors.New("is nil")

	// ErrNotImplementInterface is error triggered when provided type not implement alias interface.
	ErrNotImplementInterface = errors.New("not implement interface")

	// ErrDoesNotExist triggered when type not present in container.
	ErrDoesNotExist = errors.New("does not exist")

	// ErrMustBeSliceOrPointer triggered when resolved in invalid target.
	ErrMustBeSliceOrPointer = errors.New("must be a slice or pointer")

	// ErrMultipleDefinitions triggered when type resolved in single instance, but container contain multiple types.
	ErrMultipleDefinitions = errors.New("multiple definitions")

	// ErrorMustBeFunction triggered when value not a function.
	ErrorMustBeFunction = errors.New("must be a function")

	// ErrInvalidConstructor is error triggered when constructor have invalid signature.
	ErrInvalidConstructor = compiler.ErrInvalidConstructor

	// ErrInvalidType is error triggered when provided invalid type.
	ErrInvalidType = compiler.ErrInvalidType

	// ErrInvalidValue is error triggered when provided invalid value.
	ErrInvalidValue = compiler.ErrInvalidValue
)
