// Package metrics provides thread-safe metrics collection for the cache server.
package metrics

import (
	"sync/atomic"
)

// CacheMetrics holds cache operation counters.
type CacheMetrics struct {
	sets        uint64
	hits        uint64
	misses      uint64
	deletes     uint64
	corruptions uint64
}

// IncSet increments the set operation counter.
func (m *CacheMetrics) IncSet() {
	atomic.AddUint64(&m.sets, 1)
}

// IncHit increments the cache hit counter.
func (m *CacheMetrics) IncHit() {
	atomic.AddUint64(&m.hits, 1)
}

// IncMiss increments the cache miss counter.
func (m *CacheMetrics) IncMiss() {
	atomic.AddUint64(&m.misses, 1)
}

// IncDel increments the delete operation counter.
func (m *CacheMetrics) IncDel() {
	atomic.AddUint64(&m.deletes, 1)
}

// IncCorruption increments the data corruption counter.
func (m *CacheMetrics) IncCorruption() {
	atomic.AddUint64(&m.corruptions, 1)
}

// Get returns a snapshot of current metrics.
func (m *CacheMetrics) Get() *CacheMetrics {
	return &CacheMetrics{
		sets:        atomic.LoadUint64(&m.sets),
		hits:        atomic.LoadUint64(&m.hits),
		misses:      atomic.LoadUint64(&m.misses),
		deletes:     atomic.LoadUint64(&m.deletes),
		corruptions: atomic.LoadUint64(&m.corruptions),
	}
}

// Sets returns the total number of set operations.
func (m *CacheMetrics) Sets() uint64 {
	return atomic.LoadUint64(&m.sets)
}

// Hits returns the total number of cache hits.
func (m *CacheMetrics) Hits() uint64 {
	return atomic.LoadUint64(&m.hits)
}

// Misses returns the total number of cache misses.
func (m *CacheMetrics) Misses() uint64 {
	return atomic.LoadUint64(&m.misses)
}

// Deletes returns the total number of delete operations.
func (m *CacheMetrics) Deletes() uint64 {
	return atomic.LoadUint64(&m.deletes)
}

// Corruptions returns the total number of data corruption events.
func (m *CacheMetrics) Corruptions() uint64 {
	return atomic.LoadUint64(&m.corruptions)
}

// HitRate returns the cache hit rate as a percentage (0-100).
func (m *CacheMetrics) HitRate() float64 {
	hits := atomic.LoadUint64(&m.hits)
	misses := atomic.LoadUint64(&m.misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}
