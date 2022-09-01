// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package runtime

import (
	"fmt"
	"runtime"
	"strings"
)

type (
	// Frame is caller frame.
	Frame interface {
		// Name is function name without path and package.
		Name() string

		// File is function file name.
		File() string

		// Line is function line number.
		Line() int
	}

	frame struct {
		name string
		file string
		line int
	}
)

// frame implements the fmt.Stringer interface.
var _ fmt.Stringer = (*frame)(nil)

func Caller(skip int) Frame {
	if pc, file, line, ok := runtime.Caller(skip + 2); ok {
		return &frame{
			name: runtime.FuncForPC(pc).Name(),
			file: file,
			line: line,
		}
	}

	return &frame{
		name: "unknown",
		file: "unknown",
		line: 0,
	}
}

func (f *frame) Name() string {
	var name = f.name[strings.LastIndex(f.name, "/")+1:]
	return name[strings.Index(name, ".")+1:]
}

func (f *frame) File() string {
	return f.file
}

func (f *frame) Line() int {
	return f.line
}

func (f *frame) String() string {
	return fmt.Sprintf("%s:%d", f.File(), f.Line())
}
