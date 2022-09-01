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
	var (
		cl  = cycle.New()
		err error
	)

	require.NotPanics(t, func() {
		cl.Del(1)
	})

	err = cl.Add(1)
	require.NoError(t, err)

	err = cl.Add(1)
	require.Error(t, err)
	require.ErrorIs(t, err, cycle.ErrCycleDetected)

	cl.Del(1)
	err = cl.Add(1)
	require.NoError(t, err)
}
