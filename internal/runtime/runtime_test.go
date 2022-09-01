// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gozix/di/internal/runtime"
)

func TestCaller(t *testing.T) {
	type (
		frame struct {
			Name string
			File string
			Line int
		}

		testCase struct {
			Name  string
			Skip  int
			Frame frame
		}
	)

	var testCases = []testCase{{
		Name: "Positive case",
		Skip: 100,
		Frame: frame{
			Name: "unknown",
			File: "unknown",
			Line: 0,
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var frame = runtime.Caller(tc.Skip)

			require.Equal(t, tc.Frame.Name, frame.Name())
			require.Equal(t, tc.Frame.File, frame.File())
			require.Equal(t, tc.Frame.Line, frame.Line())
		})
	}
}
