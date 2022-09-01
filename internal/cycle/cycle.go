// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cycle

import (
	"errors"
)

// Cycle is cycle checker.
type Cycle struct {
	items map[int]bool
	stack []int
}

// ErrCycleDetected is error triggered when was cycle detected.
var ErrCycleDetected = errors.New("cycle detected")

// New is cycle constructor.
func New() *Cycle {
	return &Cycle{
		items: map[int]bool{},
		stack: []int{},
	}
}

// Add checks if same def id already exist and return error
// or mark item as visited.
func (c *Cycle) Add(key int) error {
	if c.items[key] {
		return ErrCycleDetected
	}

	c.items[key] = true
	c.stack = append(c.stack, key)

	return nil
}

// Del remove item.
func (c *Cycle) Del(key int) {
	delete(c.items, key)
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}

func (c *Cycle) Stack() []int {
	var stack = make([]int, len(c.stack))
	copy(stack, c.stack)

	return stack
}
