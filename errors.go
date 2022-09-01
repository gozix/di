// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"fmt"
	"reflect"
)

// TypeError records an error and type that caused it.
type TypeError struct {
	Type reflect.Type
	Err  error
}

// NewTypeError is error constructor.
func NewTypeError(typ reflect.Type, err error) error {
	if err == nil {
		return nil
	}

	return &TypeError{
		Type: typ,
		Err:  err,
	}
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("type %s : %s", e.Type, e.Err.Error())
}

func (e *TypeError) Unwrap() error {
	return e.Err
}
