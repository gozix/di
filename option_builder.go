// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

type builderOptionFunc func(*builder) error

// builderOptionFunc implements the BuilderOption interface.
var _ BuilderOption = (*builderOptionFunc)(nil)

func (b builderOptionFunc) applyBuilderOption(builder *builder) error {
	return b(builder)
}

// Options allow to organize logical groups of builder options.
//
//	var builder = di.NewBuilder(
//	    di.Options(
//	        di.Provide(NewBarController),
//	        di.Provide(NewBazController),
//	    ),
//	    di.Options(
//	        di.Provide(NewMuxServer),
//	        di.Provide(NewServer),
//	    ),
//	)
func Options(options ...BuilderOption) BuilderOption {
	return builderOptionFunc(func(b *builder) (err error) {
		for _, o := range options {
			if err = o.applyBuilderOption(b); err != nil {
				return err
			}
		}

		return nil
	})
}

// Add is builder constructor option.
// This is a syntax sugar for builder constructor usage.
func Add(value Value, options ...AddOption) BuilderOption {
	var option = caller(1)
	return builderOptionFunc(func(b *builder) error {
		return b.Add(value, append([]AddOption{option}, options...)...)
	})
}

// Autowire is builder constructor option.
// This is a syntax sugar for builder constructor usage.
func Autowire(value Type, options ...ProvideOption) BuilderOption {
	var option = caller(1)
	return builderOptionFunc(func(b *builder) error {
		return b.Autowire(value, append([]ProvideOption{option}, options...)...)
	})
}

// Provide is builder constructor option.
// This is a syntax sugar for builder constructor usage.
func Provide(value Constructor, options ...ProvideOption) BuilderOption {
	var option = caller(1)
	return builderOptionFunc(func(b *builder) error {
		return b.Provide(value, append([]ProvideOption{option}, options...)...)
	})
}
