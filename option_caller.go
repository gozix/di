// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di

import "github.com/gozix/di/internal/runtime"

type callerOption struct {
	frame runtime.Frame
}

// callerOption implements the AsOption interface.
var _ AddOption = (*callerOption)(nil)

func caller(skip int) *callerOption {
	return &callerOption{
		frame: runtime.Caller(skip),
	}
}

func (o *callerOption) applyAddOption(def *definition) {
	def.frame = o.frame
}

func (o *callerOption) applyProvideOption(def *definition) {
	def.frame = o.frame
}
