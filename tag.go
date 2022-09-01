// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type (
	// Arg is representation.
	Arg struct {
		Key   string
		Value string
	}

	// Args is arg collection.
	Args []Arg

	// Tag is tag representation.
	Tag struct {
		Name string
		Args Args
	}

	// Tags is tag collection.
	Tags []Tag
)

var (
	// Tags implements the AddOption interface.
	_ AddOption = (*Tags)(nil)

	// Tags implements the ProvideOption interface.
	_ ProvideOption = (*Tags)(nil)
)

func (t Tags) applyAddOption(def *definition) {
	def.tags = append(def.tags, t...)
}

func (t Tags) applyProvideOption(def *definition) {
	def.tags = append(def.tags, t...)
}

func (t Tags) contains(name string) bool {
	for _, t1 := range t {
		if t1.Name == name {
			return true
		}
	}

	return false
}
