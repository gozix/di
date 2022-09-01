// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package compiler

import (
	"errors"
	"reflect"
)

type Type struct {
	typ reflect.Type
}

var (
	// Type implements the Compiler interface.
	_ Compiler = (*Type)(nil)

	// ErrInvalidType is error triggered when provided invalid type.
	ErrInvalidType = errors.New("invalid type")
)

func NewType(v any) (*Type, error) {
	var rt = reflect.TypeOf(v)
	if rt == nil || rt.Kind() == reflect.Invalid {
		return nil, ErrInvalidType
	}

	return &Type{typ: rt}, nil
}

func (c *Type) Create(deps ...*Dependency) (reflect.Value, Closer, error) {
	if c.typ.Kind() != reflect.Ptr {
		return reflect.New(c.typ).Elem(), nil, nil
	}

	var (
		rt = c.typ.Elem()
		rz = reflect.Zero(rt)
		rv = reflect.New(rt)
	)

	rv.Elem().Set(rz)

	for _, dep := range deps {
		rv.Elem().Field(dep.Index).Set(dep.Value)
	}

	return rv, nil, nil
}

func (c *Type) Dependencies() []*Dependency {
	var rt = c.typ
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return nil
	}

	var deps = make([]*Dependency, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).IsExported() {
			deps = append(deps, &Dependency{
				Name:  rt.Field(i).Name,
				Index: i,
				Type:  rt.Field(i).Type,
				Value: reflect.New(rt.Field(i).Type).Elem(),
			})
		}
	}

	return deps
}

func (c *Type) Type() reflect.Type {
	return c.typ
}
