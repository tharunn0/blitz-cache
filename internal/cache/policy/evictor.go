// Package policy defines eviction policy interface.
package policy

import (
	"cache-server/pkg/types"
)

type Evictor interface {
	ShouldEvict(count, size int64) bool
	Evict(items map[string]*Entry) // mutates map, removes entries
}

type Entry = types.Entry // alias for convenience
