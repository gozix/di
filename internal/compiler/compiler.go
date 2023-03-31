// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package compiler

import (
	"errors"
	"reflect"
)

type (
	Closer = func() error

	Compiler interface {
		Create(dependencies ...*Dependency) (reflect.Value, Closer, error)
		Dependencies() []*Dependency
		Type() reflect.Type
	}

	Dependency struct {
		Name  string
		Index int
		Type  reflect.Type
		Value reflect.Value
	}
)

var ErrPanicked = errors.New("panicked")
