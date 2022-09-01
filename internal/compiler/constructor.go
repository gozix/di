// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package compiler

import (
	"errors"
	"fmt"
	"reflect"
)

// Constructor implements the Compiler interface.
type Constructor struct {
	typ reflect.Type
	val reflect.Value
	vct bool
	lin int
	beh int
}

const (
	behaviourUnknown = iota
	behaviourValue
	behaviourValueError
	behaviourValueCloser
	behaviourValueCloserError
)

var (
	// compile time check.
	_ Compiler = (*Constructor)(nil)

	// reflectErrorType is error reflect type cache.
	reflectErrorType = reflect.TypeOf((*error)(nil)).Elem()

	// ErrInvalidConstructor is error triggered when constructor have invalid signature.
	ErrInvalidConstructor = errors.New("unexpected constructor")
)

// NewConstructor is constructor of Constructor.
// Argument fn may be one of:
//   - func New(args ...any) (value any)
//   - func New(args ...any) (value any, err error)
//   - func New(args ...any) (value any, closer func(){})
//   - func New(args ...any) (value any, closer func(){}, err error)
func NewConstructor(fn any) (*Constructor, error) {
	var c = &Constructor{
		typ: reflect.TypeOf(fn),
		val: reflect.ValueOf(fn),
	}

	c.guessBehaviour()
	if c.beh == behaviourUnknown {
		return nil, fmt.Errorf("got %v : %w", c.typ, ErrInvalidConstructor)
	}

	c.vct = c.typ.IsVariadic()
	c.lin = c.typ.NumIn() - 1

	return c, nil
}

func (c *Constructor) Create(dependencies ...*Dependency) (reflect.Value, Closer, error) {
	var args = make([]reflect.Value, 0, len(dependencies))
	for i, dep := range dependencies {
		if c.vct && i == c.lin {
			for i = 0; i < dep.Value.Len(); i++ {
				args = append(args, dep.Value.Index(i))
			}

			continue
		}

		args = append(args, dep.Value)
	}

	var out = c.val.Call(args)
	switch c.beh {
	case behaviourValue:
		return out[0], nil, nil
	case behaviourValueError:
		return out[0], nil, c.nilOrError(out[1])
	case behaviourValueCloser:
		return out[0], c.nilOrCloser(out[1]), nil
	case behaviourValueCloserError:
		return out[0], c.nilOrCloser(out[1]), c.nilOrError(out[2])
	}

	return reflect.Value{}, nil, ErrInvalidConstructor
}

func (c *Constructor) Dependencies() []*Dependency {
	var deps = make([]*Dependency, c.typ.NumIn())
	for i := 0; i < c.typ.NumIn(); i++ {
		deps[i] = &Dependency{
			Name:  c.typ.In(i).Name(),
			Index: i,
			Type:  c.typ.In(i),
			Value: reflect.New(c.typ.In(i)).Elem(),
		}
	}

	return deps
}

func (c *Constructor) Type() reflect.Type {
	return c.typ.Out(0)
}

func (c *Constructor) guessBehaviour() {
	switch {
	case c.typ == nil:
		c.beh = behaviourUnknown
	case c.typ.Kind() != reflect.Func:
		c.beh = behaviourUnknown
	case c.typ.NumOut() == 1:
		c.beh = behaviourValue
	case c.typ.NumOut() == 2 && c.isError(c.typ.Out(1)):
		c.beh = behaviourValueError
	case c.typ.NumOut() == 2 && c.isCloser(c.typ.Out(1)):
		c.beh = behaviourValueCloser
	case c.typ.NumOut() == 3 && c.isCloser(c.typ.Out(1)) && c.isError(c.typ.Out(2)):
		c.beh = behaviourValueCloserError
	}
}

func (*Constructor) isError(t reflect.Type) bool {
	return t.Implements(reflectErrorType)
}

func (c *Constructor) isCloser(t reflect.Type) bool {
	return t.Kind() == reflect.Func && t.NumIn() == 0 && t.NumOut() == 1 && c.isError(t.Out(0))
}

func (*Constructor) nilOrCloser(v reflect.Value) Closer {
	return v.Interface().(Closer)
}

func (*Constructor) nilOrError(v reflect.Value) error {
	if v.IsNil() {
		return nil
	}

	return v.Interface().(error)
}
