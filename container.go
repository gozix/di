// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/gozix/di/internal/compiler"
	"github.com/gozix/di/internal/cycle"
	"github.com/gozix/di/internal/runtime"
)

type (
	// container cache
	cache map[int]reflect.Value

	// container implements the Container interface.
	container struct {
		mux     sync.Mutex
		defs    definitions
		cache   map[int]reflect.Value
		closers []compiler.Closer
	}
)

var (
	// container implements the Container interface.
	_ Container = (*container)(nil)

	// reflectErrorType is error reflect type cache.
	reflectErrorType = reflect.TypeOf((*error)(nil)).Elem()

	// reflectContainerType is Container reflect type cache.
	reflectContainerType = reflect.TypeOf((*Container)(nil)).Elem()
)

func (c *container) Call(fn Function, options ...ConstraintOption) (err error) {
	var rv = reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		return fmt.Errorf("%s : fn %w", runtime.Caller(0), ErrorMustBeFunction)
	}

	var (
		rt = rv.Type()
		in = make([]reflect.Value, 0, rt.NumIn())
		cs = make(constraints, len(options))
	)

	for _, o := range options {
		o.applyConstraintOption(cs)
	}

	var (
		vt = rt.IsVariadic()
		ic = rt.NumIn() - 1
	)

	for i := 0; i <= ic; i++ {
		var (
			at  = rt.In(i)
			dep = &compiler.Dependency{
				Name:  at.Name(),
				Index: i,
				Type:  at,
				Value: reflect.New(at),
			}
		)

		c.mux.Lock()
		err = c.resolveDependency(cycle.New(), dep, cs)
		c.mux.Unlock()

		if err != nil {
			return fmt.Errorf("%s : %w", runtime.Caller(0), err)
		}

		if vt && i == ic {
			for i = 0; i < dep.Value.Elem().Len(); i++ {
				in = append(in, dep.Value.Elem().Index(i))
			}

			continue
		}

		in = append(in, dep.Value.Elem())
	}

	var out = rv.Call(in)
	if len(out) > 0 && out[len(out)-1].Type().Implements(reflectErrorType) && !out[rt.NumOut()-1].IsNil() {
		return fmt.Errorf("%s : %w", runtime.Caller(0), out[rt.NumOut()-1].Interface().(error))
	}

	return nil
}

func (c *container) Close() (err error) {
	c.mux.Lock()

	var closers = append(([]compiler.Closer)(nil), c.closers...)
	c.defs = make(definitions, 0)
	c.closers = c.closers[:0]

	c.mux.Unlock()

	for i := len(closers) - 1; i >= 0; i-- {
		if err = closers[i](); err != nil {
			return fmt.Errorf("unable to close container : %w", err)
		}
	}

	return nil
}

func (c *container) Has(value Type, modifiers ...Modifier) bool {
	var (
		rt   = reflect.TypeOf(value)
		defs = c.defs.find(rt, modifiers)
	)

	return len(defs) > 0
}

func (c *container) Resolve(target Value, modifiers ...Modifier) (err error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var rv = reflect.ValueOf(target)
	if err = c.resolve(&rv, cycle.New(), modifiers); err != nil {
		return fmt.Errorf("%s : %w", runtime.Caller(0), err)
	}

	return nil
}

func (c *container) resolve(tv *reflect.Value, cycle *cycle.Cycle, modifiers []Modifier) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("unable to resolve target because the container panicked: %+v", recovered)
		}
	}()

	if tv.Kind() != reflect.Pointer && tv.Kind() != reflect.Slice {
		if tv.IsValid() {
			return NewTypeError(tv.Type(), ErrMustBeSliceOrPointer)
		}

		return ErrMustBeSliceOrPointer
	}

	var ft = tv.Type().Elem()
	if reflectContainerType.AssignableTo(ft) {
		c.set(tv, reflect.ValueOf(c))
		return
	}

	var defs = c.defs.find(ft, modifiers)
	if len(defs) == 0 {
		if ft.Kind() == reflect.Slice {
			defs = c.defs.find(ft.Elem(), modifiers)
		}

		if len(defs) == 0 {
			return NewTypeError(tv.Type(), ErrDoesNotExist)
		}
	}

	if ft.Kind() != reflect.Slice && len(defs) > 1 {
		return NewTypeError(tv.Type(), ErrMultipleDefinitions)
	}

	for _, def := range defs {
		if err = cycle.Add(def.id); err != nil {
			return NewTypeError(tv.Type(), err)
		}

		if !def.unshared {
			var rv = c.cache[def.id]
			if rv.IsValid() {
				c.set(tv, rv)
				cycle.Del(def.id)
				continue
			}
		}

		var deps = def.compiler.Dependencies()
		for _, dep := range deps {
			if err = c.resolveDependency(cycle, dep, def.constraints); err != nil {
				cycle.Del(def.id)
				return err
			}
		}

		var sv, closer, err = def.compiler.Create(deps...)
		if err != nil {
			cycle.Del(def.id)
			return NewTypeError(def.compiler.Type(), err)
		}

		if !def.unshared {
			c.cache[def.id] = sv
		}

		if closer != nil {
			c.closers = append(c.closers, closer)
		}

		c.set(tv, sv)
		cycle.Del(def.id)
	}

	return nil
}

func (c *container) resolveDependency(cycle *cycle.Cycle, dep *compiler.Dependency, cs constraints) error {
	var v = &dep.Value
	if v.CanAddr() {
		v = &[]reflect.Value{v.Addr()}[0]
	}

	var (
		constr = cs.choose(dep.Index, dep.Name, dep.Type)
		err    = c.resolve(v, cycle, constr.modifiers)
	)

	if errors.Is(err, ErrDoesNotExist) && constr.optional {
		if e, ok := err.(*TypeError); ok && v.Type() != e.Type {
			return err
		}

		return nil
	}

	return err
}

func (c *container) set(tv *reflect.Value, sv reflect.Value) {
	switch tv.Elem().Kind() {
	case reflect.Slice:
		if sv.Kind() == reflect.Slice {
			tv.Elem().Set(
				reflect.AppendSlice(tv.Elem(), sv),
			)
		} else {
			tv.Elem().Set(
				reflect.Append(tv.Elem(), sv),
			)
		}
	default:
		tv.Elem().Set(sv)
	}
}
