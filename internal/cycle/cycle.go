// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cycle

// Cycle is cycle checker.
type Cycle struct {
	items map[int]bool
}

// New is cycle constructor.
func New() *Cycle {
	return &Cycle{
		items: map[int]bool{},
	}
}

// Append creates a new chain and adds the key to it
func (c *Cycle) Append(key int) *Cycle {
	var clone = &Cycle{
		items: map[int]bool{},
	}

	for k, v := range c.items {
		clone.items[k] = v
	}

	clone.items[key] = true

	return clone
}

// Has return true if the key exists
func (c *Cycle) Has(key int) bool {
	return c.items[key]
}
