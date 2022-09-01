// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package compiler

import (
	"errors"
	"reflect"
)

type Value struct {
	rv reflect.Value
}

var (
	// Value implements the Compiler interface.
	_ Compiler = (*Value)(nil)

	// ErrInvalidValue is error triggered when provided invalid value.
	ErrInvalidValue = errors.New("invalid value")
)

func NewValue(v any) (*Value, error) {
	var rv = reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, ErrInvalidValue
	}

	return &Value{rv: rv}, nil
}

func (c *Value) Create(_ ...*Dependency) (reflect.Value, Closer, error) {
	return c.rv, nil, nil
}

func (c *Value) Dependencies() []*Dependency {
	return nil
}

func (c *Value) Type() reflect.Type {
	return c.rv.Type()
}
