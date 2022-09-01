// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type (
	// BuilderOption is specified for NewBuilder option interface.
	BuilderOption interface {
		applyBuilderOption(*builder) error
	}

	// AddOption is specified for Builder.Add method option interface.
	AddOption interface {
		applyAddOption(def *definition)
	}

	// ProvideOption is specified for Builder.Provide method option interface.
	ProvideOption interface {
		applyProvideOption(def *definition)
	}
)
