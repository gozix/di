// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cycle_test

import (
	"testing"

	"github.com/gozix/di/internal/cycle"
	"github.com/stretchr/testify/require"
)

func TestCycle(t *testing.T) {
	var cl = cycle.New()

	require.False(t, cl.Has(1))

	var c2 = cl.Append(1)

	require.False(t, cl.Has(1))
	require.True(t, c2.Has(1))
}
