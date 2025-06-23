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
	// container implements the Container interface.
	container struct {
		*containerCore
		*resolver

		cycle *cycle.Cycle
	}

	// container core values
	containerCore struct {
		mux     sync.Mutex
		defs    definitions
		cache   cache
		closers []compiler.Closer
	}

	// container dependency resolver
	resolver struct{}

	// container cache
	cache map[int]*cacheItem

	// container cache item
	cacheItem struct {
		value reflect.Value
		ready chan struct{}
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

		err = c.resolveDependency(c, dep, cs)

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
	c.mux.Lock()
	defer c.mux.Unlock()

	var (
		rt   = reflect.TypeOf(value)
		defs = c.defs.find(rt, modifiers)
	)

	return len(defs) > 0
}

func (c *container) Resolve(target Value, modifiers ...Modifier) (err error) {
	var rv = reflect.ValueOf(target)
	if err = c.resolve(c, &rv, modifiers); err != nil {
		return fmt.Errorf("%s : %w", runtime.Caller(0), err)
	}

	return nil
}

func (r *resolver) resolve(ctn *container, tv *reflect.Value, modifiers []Modifier) (err error) {
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
		r.set(tv, reflect.ValueOf(ctn))
		return
	}

	ctn.mux.Lock()
	var defs = ctn.defs.find(ft, modifiers)
	ctn.mux.Unlock()

	if len(defs) == 0 {
		if ft.Kind() == reflect.Slice {
			ctn.mux.Lock()
			defs = ctn.defs.find(ft.Elem(), modifiers)
			ctn.mux.Unlock()
		}

		if len(defs) == 0 {
			return NewTypeError(tv.Type(), ErrDoesNotExist)
		}
	}

	if ft.Kind() != reflect.Slice && len(defs) > 1 {
		return NewTypeError(tv.Type(), ErrMultipleDefinitions)
	}

	for _, def := range defs {
		if ctn.cycle.Has(def.id) {
			return NewTypeError(tv.Type(), ErrCycleDetected)
		}

		if !def.unshared {
			ctn.mux.Lock()
			var dep, ok = ctn.cache[def.id]
			ctn.mux.Unlock()

			if ok {
				<-dep.ready
				r.set(tv, dep.value)
				continue
			}

			ctn.mux.Lock()
			ctn.cache[def.id] = &cacheItem{
				ready: make(chan struct{}),
			}
			ctn.mux.Unlock()
		}

		var deps = def.compiler.Dependencies()
		for _, dep := range deps {
			var newCtn = &container{
				containerCore: ctn.containerCore,
				cycle:         ctn.cycle.Append(def.id),
			}

			if err = r.resolveDependency(newCtn, dep, def.constraints); err != nil {
				return err
			}
		}

		var sv, closer, err = def.compiler.Create(deps...)
		if err != nil {
			return NewTypeError(def.compiler.Type(), err)
		}

		if !def.unshared {
			ctn.mux.Lock()
			ctn.cache[def.id].value = sv
			close(ctn.cache[def.id].ready)
			ctn.mux.Unlock()
		}

		if closer != nil {
			ctn.mux.Lock()
			ctn.closers = append(ctn.closers, closer)
			ctn.mux.Unlock()
		}

		r.set(tv, sv)
	}

	return nil
}

func (r *resolver) resolveDependency(ctn *container, dep *compiler.Dependency, cs constraints) error {
	var v = &dep.Value
	if v.CanAddr() {
		v = &[]reflect.Value{v.Addr()}[0]
	}

	var (
		constr = cs.choose(dep.Index, dep.Name, dep.Type)
		err    = ctn.resolve(ctn, v, constr.modifiers)
	)

	if errors.Is(err, ErrDoesNotExist) && constr.optional {
		if e, ok := err.(*TypeError); ok && v.Type() != e.Type {
			return err
		}

		return nil
	}

	return err
}

func (r *resolver) set(tv *reflect.Value, sv reflect.Value) {
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
