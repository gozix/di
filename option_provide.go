// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type (
	// ProvideOption is specified for Builder.Provide method option interface.
	ProvideOption interface {
		applyProvideOption(def *definition)
	}

	provideOptionFunc func(def *definition)
)

// provideOptionFunc implements the AddOption interface.
var _ ProvideOption = (*provideOptionFunc)(nil)

func (fn provideOptionFunc) applyProvideOption(def *definition) {
	fn(def)
}

// ProvideOptions allow to organize logical groups of provide options.
//
//	var builder = di.NewBuilder()
//	builder.Provide(di.ProvideOptions(
//		di.As(new(io.Writer)),
//		di.Tags{{
//			Name: "tag",
//		}},
//	))
func ProvideOptions(options ...ProvideOption) ProvideOption {
	return provideOptionFunc(func(def *definition) {
		for _, o := range options {
			o.applyProvideOption(def)
		}
	})
}
