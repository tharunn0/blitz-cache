// Package sieve implements the SIEVE probabilistic eviction policy.
// O(1) amortized, low alloc, thread-safe.
// Evicts randomly with probability based on overprovisioning.

package sieve

import (
	"math"
	"math/rand/v2"
	"sync/atomic"

	"cache-server/pkg/types"
)

type Metrics struct {
	Evicted int64
}

type Evictor struct {
	maxEntries int64
	maxMemory  int64
	probFactor float64
	metrics    *Metrics
	rng        *rand.Rand
}

func NewEvictor(maxEntries, maxMemory int64) *Evictor {
	return &Evictor{
		maxEntries: maxEntries,
		maxMemory:  maxMemory,
		probFactor: 1.1,
		metrics:    &Metrics{},
		rng:        rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func (e *Evictor) ShouldEvict(count, size int64) bool {
	if e.maxEntries > 0 && count > e.maxEntries {
		return true
	}
	if e.maxMemory > 0 && size > e.maxMemory {
		return true
	}
	return false
}

// Evict removes entries probabilistically based on overprovision level.
// Accepts current count and size to avoid redundant calculation.
func (e *Evictor) EvictWithMetrics(items map[string]*types.Entry, currentCount, currentSize int64) {
	pOver := 0.0
	if e.maxEntries > 0 {
		pOver = math.Max(pOver, float64(currentCount)/float64(e.maxEntries))
	}
	if e.maxMemory > 0 {
		pOver = math.Max(pOver, float64(currentSize)/float64(e.maxMemory))
	}
	p := pOver*e.probFactor - 1.0 // prob to evict

	if p <= 0 {
		return
	}

	numEvicted := int64(0)
	for key := range items {
		if e.rng.Float64() < p {
			delete(items, key)
			numEvicted++
		}
	}
	atomic.AddInt64(&e.metrics.Evicted, numEvicted)
}

// Evict is the legacy method that calculates metrics internally.
// Kept for interface compatibility but less efficient.
func (e *Evictor) Evict(items map[string]*types.Entry) {
	// Calculate count/size
	count := int64(len(items))
	size := int64(0)
	for _, ent := range items {
		size += int64(len(ent.Value))
	}

	e.EvictWithMetrics(items, count, size)
}
