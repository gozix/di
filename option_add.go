// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type (
	// AddOption is specified for Builder.Add method option interface.
	AddOption interface {
		applyAddOption(def *definition)
	}

	addOptionFunc func(def *definition)
)

// provideOptionFunc implements the AddOption interface.
var _ AddOption = (*addOptionFunc)(nil)

func (fn addOptionFunc) applyAddOption(def *definition) {
	fn(def)
}

// AddOptions allow to organize logical groups of add options.
//
//	var builder = di.NewBuilder()
//	builder.Add(di.AddOptions(
//		di.As(new(io.Writer)),
//		di.Tags{{
//			Name: "tag",
//		}},
//	))
func AddOptions(options ...AddOption) AddOption {
	return addOptionFunc(func(def *definition) {
		for _, o := range options {
			o.applyAddOption(def)
		}
	})
}
