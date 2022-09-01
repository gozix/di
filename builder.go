// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/gozix/di/internal/compiler"
	"github.com/gozix/di/internal/runtime"
)

// builder implements the Builder interface.
type builder struct {
	defs definitions
	mux  sync.Mutex
	seq  int
}

// NewBuilder is builder constructor.
func NewBuilder(options ...BuilderOption) (_ Builder, err error) {
	var b = &builder{
		defs: definitions{},
	}

	return b, b.Apply(options...)
}

func (b *builder) Add(value Value, options ...AddOption) (err error) {
	var def = &definition{
		constraints: constraints{},
	}

	def.applyAddOptions(options...)
	if def.frame == nil {
		def.frame = runtime.Caller(0)
	}

	if def.compiler, err = compiler.NewValue(value); err != nil {
		return fmt.Errorf("%s : %w", def.frame, err)
	}

	return b.add(def)
}

func (b *builder) Apply(options ...BuilderOption) (err error) {
	for _, opt := range options {
		if err = opt.applyBuilderOption(b); err != nil {
			return err
		}
	}

	return nil
}

func (b *builder) Autowire(value Type, options ...ProvideOption) (err error) {
	var def = &definition{
		constraints: constraints{},
	}

	def.applyProvideOptions(options...)
	if def.frame == nil {
		def.frame = runtime.Caller(0)
	}

	if def.compiler, err = compiler.NewType(value); err != nil {
		return fmt.Errorf("%s : %w", def.frame, err)
	}

	return b.add(def)
}

func (b *builder) Provide(value Constructor, options ...ProvideOption) (err error) {
	var def = &definition{
		constraints: constraints{},
	}

	def.applyProvideOptions(options...)
	if def.frame == nil {
		def.frame = runtime.Caller(0)
	}

	if def.compiler, err = compiler.NewConstructor(value); err != nil {
		return fmt.Errorf("%s : %w", def.frame, err)
	}

	return b.add(def)
}

func (b *builder) Build() (Container, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	var defs = definitions{}
	for k, v := range b.defs {
		defs[k] = v
	}

	return &container{
		defs:  defs,
		cache: cache{},
	}, nil
}

func (b *builder) Definitions() []Definition {
	b.mux.Lock()
	defer b.mux.Unlock()

	var defs = make([]Definition, 0, len(b.defs))
	for i := range b.defs {
		for j := range b.defs[i] {
			var def = b.defs[i][j]
			def.definitions = b.defs

			defs = append(defs, Definition(&def))
		}
	}

	return defs
}

func (b *builder) add(def *definition) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.seq++
	def.id = b.seq

	var ct = def.compiler.Type()
	b.defs[ct] = append(b.defs[ct], *def)

	for _, alias := range def.aliases {
		if alias == nil {
			return fmt.Errorf("%s : %w", def.frame, ErrIsNil)
		}

		var at = reflect.TypeOf(alias)
		if at.Kind() != reflect.Ptr || at.Elem().Kind() != reflect.Interface {
			return fmt.Errorf("%s : %w", def.frame, NewTypeError(at, ErrNotPointerToInterface))
		}

		if ct == at.Elem() {
			continue
		}

		if !ct.Implements(at.Elem()) {
			return fmt.Errorf("%s : %s %w %s", def.frame, ct.String(), ErrNotImplementInterface, at.String())
		}

		b.defs[at.Elem()] = append(b.defs[at.Elem()], *def)
	}

	return nil
}
