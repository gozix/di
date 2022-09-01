// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

// unsharedOption is an option
type unsharedOption struct {
	value bool
}

// unsharedOption implements the ProvideOption interface.
var _ ProvideOption = (*unsharedOption)(nil)

// Unshared mark definition as Unshared.
func Unshared() ProvideOption {
	return &unsharedOption{value: true}
}

func (o *unsharedOption) applyProvideOption(def *definition) {
	def.unshared = o.value
}
