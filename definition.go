// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"reflect"

	"github.com/gozix/di/internal/compiler"
	"github.com/gozix/di/internal/runtime"
)

type (
	// definition is container item representation.
	definition struct {
		id          int
		aliases     []any
		compiler    compiler.Compiler
		constraints constraints
		frame       runtime.Frame
		tags        Tags
		unshared    bool

		definitions definitions
	}

	// definitions are list of definitions.
	definitions map[reflect.Type][]definition
)

// definition implements the Definition interface.
var _ Definition = (*definition)(nil)

func (d *definition) Dependencies() []Dependency {
	var deps []Dependency
	for _, dep := range d.compiler.Dependencies() {
		var (
			constr = d.constraints.choose(dep.Index, dep.Name, dep.Type)
			defs   = make([]Definition, 0, 2)
		)

		for _, def := range d.definitions.find(dep.Type, constr.modifiers) {
			def.definitions = d.definitions
			defs = append(defs, Definition(&def))
		}

		deps = append(deps, Dependency{
			Type:        dep.Type,
			Optional:    constr.optional,
			Definitions: defs,
		})
	}

	return deps
}

func (d *definition) ID() int {
	return d.id
}

func (d *definition) Type() reflect.Type {
	return d.compiler.Type()
}

func (d *definition) Tags() Tags {
	var tags = make(Tags, len(d.tags))
	copy(tags, d.tags)

	return tags
}

func (d *definition) Unshared() bool {
	return d.unshared
}

func (d *definition) applyAddOptions(options ...AddOption) {
	for _, o := range options {
		o.applyAddOption(d)
	}
}

func (d *definition) applyProvideOptions(options ...ProvideOption) {
	for _, o := range options {
		o.applyProvideOption(d)
	}
}

func (d definitions) find(typ reflect.Type, modifiers []Modifier) (founded []definition) {
	var defs = make([]Definition, 0, 4)
	for i := range d[typ] {
		defs = append(defs, &d[typ][i])
	}

	for _, mod := range modifiers {
		defs = mod(defs)
	}

	for _, def := range defs {
		if def == nil {
			continue
		}

		founded = append(founded, *def.(*definition))
	}

	return founded
}
