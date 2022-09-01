// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import (
	"reflect"
	"sort"
)

type (
	// ConstraintOption is an option
	ConstraintOption interface {
		ProvideOption
		applyConstraintOption(options constraints)
	}

	// Modifier calling with definitions founded by type in container. Modifiers should use for filtering,
	// sorting or any other modifications list of definitions before they should resolve.
	Modifier func(defs []Definition) []Definition

	// Restriction is an option.
	Restriction interface {
		applyRestriction(options *constraintOption)
	}

	constraint struct {
		optional  bool
		modifiers []Modifier
	}

	constraintOption struct {
		key       any
		optional  bool
		modifiers []Modifier
	}

	constraints map[any]constraint

	constraintRestrictionFunc func(*constraintOption)
)

var (
	// constraintOption implements the ConstraintOption interface.
	_ ConstraintOption = (*constraintOption)(nil)

	// constraintRestrictionFunc implements the Restriction interface.
	_ Restriction = (*constraintRestrictionFunc)(nil)

	// Modifier implements the Restriction interface.
	_ Restriction = (*Modifier)(nil)
)

// Constraint restricts the dependency resolving.
//
// The key argument may be string, int or any other type
func Constraint(key any, restrictions ...Restriction) ConstraintOption {
	var opt = new(constraintOption)
	for _, c := range restrictions {
		c.applyRestriction(opt)
	}

	switch rt := reflect.TypeOf(key); rt.Kind() {
	case reflect.String, reflect.Int:
		opt.key = key
	default:
		opt.key = rt
	}

	return opt
}

// Optional set dependency optional or not, by default all dependencies are required.
func Optional(v bool) Restriction {
	return constraintRestrictionFunc(func(option *constraintOption) {
		option.optional = v
	})
}

// Filter filters the definitions the provided match function.
func Filter(match func(def Definition) bool) Modifier {
	return func(defs []Definition) []Definition {
		var index = 0
		for _, def := range defs {
			if match(def) {
				defs[index] = def
				index++
			}
		}

		return defs[:index]
	}
}

// Sort sorts the definitions the provided less function.
func Sort(less func(a, b Definition) bool) Modifier {
	return func(defs []Definition) []Definition {
		sort.Slice(defs, func(i, j int) bool {
			return less(defs[i], defs[j])
		})

		return defs
	}
}

// WithTags filters out definitions without needed tags.
func WithTags(tags ...string) Modifier {
	return Filter(func(def Definition) bool {
		var found = true
		for _, t := range tags {
			if def.Tags().contains(t) {
				continue
			}

			found = false
		}

		return found
	})
}

// WithoutTags filters out definitions with needed tags.
func WithoutTags(tags ...string) Modifier {
	return Filter(func(def Definition) bool {
		var found = false
		for _, t := range tags {
			if def.Tags().contains(t) {
				found = true
				break
			}
		}

		return !found
	})
}

func (m Modifier) applyRestriction(options *constraintOption) {
	options.modifiers = append(options.modifiers, m)
}

func (o *constraintOption) applyConstraintOption(cs constraints) {
	cs[o.key] = constraint{
		optional:  o.optional,
		modifiers: o.modifiers,
	}
}

func (o *constraintOption) applyProvideOption(def *definition) {
	def.constraints[o.key] = constraint{
		optional:  o.optional,
		modifiers: o.modifiers,
	}
}

func (cs constraints) choose(index int, name string, typ reflect.Type) constraint {
	if _, ok := cs[index]; ok {
		return cs[index]
	}

	if _, ok := cs[name]; ok {
		return cs[name]
	}

	if _, ok := cs[typ]; ok {
		return cs[typ]
	}

	return constraint{}
}

func (fn constraintRestrictionFunc) applyRestriction(option *constraintOption) {
	fn(option)
}
