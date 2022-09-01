// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type (
	// AsOption is an option
	AsOption interface {
		AddOption
		ProvideOption
	}

	asOption struct {
		aliases []any
	}
)

// asOption implements the AsOption interface.
var _ AddOption = (*asOption)(nil)

// As sets type alias.
func As(aliases ...any) AsOption {
	return &asOption{
		aliases: aliases,
	}
}

func (o *asOption) apply(def *definition) {
	def.aliases = append(def.aliases, o.aliases...)
}

func (o *asOption) applyAddOption(def *definition) {
	o.apply(def)
}

func (o *asOption) applyProvideOption(def *definition) {
	o.apply(def)
}
