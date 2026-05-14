// Package sharded implements a sharded in-memory store with lock striping.
package sharded

import (
	"log"
	"sync"

	"cache-server/internal/cache/policy"
	"cache-server/internal/cache/policy/sieve"
	"cache-server/internal/clock"
	"cache-server/pkg/types"

	"github.com/cespare/xxhash/v2"
)

const NumShards = 256

// Compile-time check that NumShards is a power of 2
// If NumShards is a power of 2, (NumShards & (NumShards - 1)) == 0
// This creates a zero-sized array type, which is valid only if the expression is 0
type _ [NumShards & (NumShards - 1)]struct{}

type Shard struct {
	mu     sync.RWMutex
	items  map[string]*types.Entry
	size   int64 // current size in bytes
	count  int64 // number of entries
	policy policy.Evictor
}

func NewShard(p policy.Evictor) *Shard {
	return &Shard{
		items:  make(map[string]*types.Entry, 1024), // preallocate
		policy: p,
	}
}

func (s *Shard) Set(key string, value []byte, expireTick uint64) {
	s.mu.Lock()

	defer s.mu.Unlock()
	oldEntry, exists := s.items[key]
	if exists {
		s.size -= int64(len(oldEntry.Value))
		s.count--
	}

	entry := &types.Entry{Value: value, ExpireTick: expireTick}
	s.items[key] = entry
	s.size += int64(len(value))
	s.count++

	// Check eviction
	if s.policy.ShouldEvict(s.count, s.size) {
		// Use optimized eviction with current metrics
		if evictorWithMetrics, ok := s.policy.(*sieve.Evictor); ok {
			evictorWithMetrics.EvictWithMetrics(s.items, s.count, s.size)
		} else {
			s.policy.Evict(s.items)
		}
		// Update size/count after eviction
		s.updateSize()
	}
}

func (s *Shard) Get(key string, nowTick uint64) ([]byte, bool) {
	s.mu.RLock()
	entry, exists := s.items[key]
	if !exists {
		s.mu.RUnlock()
		return nil, false
	}

	// Check expiration - use < instead of <= to fix off-by-one
	if entry.ExpireTick < nowTick {
		s.mu.RUnlock()
		// Upgrade to write lock for lazy deletion
		s.mu.Lock()
		// Double-check entry still exists and is expired
		if entry, exists := s.items[key]; exists && entry.ExpireTick < nowTick {
			delete(s.items, key)
			s.size -= int64(len(entry.Value))
			s.count--
		}
		s.mu.Unlock()
		return nil, false
	}

	value := entry.Value
	s.mu.RUnlock()
	return value, true
}

func (s *Shard) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.items[key]
	if exists {
		s.size -= int64(len(entry.Value))
		s.count--
		delete(s.items, key)
	}
}

func (s *Shard) Janitor(nowTick uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, e := range s.items {
		if e.ExpireTick < nowTick { // Fixed: use < instead of <=
			s.size -= int64(len(e.Value))
			s.count--
			delete(s.items, k)
		}
	}
}

func (s *Shard) updateSize() {
	s.size = 0
	s.count = 0
	for _, e := range s.items {
		s.size += int64(len(e.Value))
		s.count++
	}
}

type Store struct {
	shards [NumShards]*Shard
	clock  *clock.Clock
}

// NewStore creates a new sharded store with per-shard evictors.
func NewStore(clock *clock.Clock, maxEntries, maxMemory int64) *Store {
	s := &Store{clock: clock}

	// Divide limits across shards
	perShardEntries := maxEntries / NumShards
	perShardMemory := maxMemory / NumShards

	for i := 0; i < NumShards; i++ {
		evictor := sieve.NewEvictor(perShardEntries, perShardMemory)
		s.shards[i] = NewShard(evictor)
	}
	return s
}

func (st *Store) shardIndex(key string) int {
	h := xxhash.Sum64String(key)
	return int(h & (NumShards - 1))
}

func (st *Store) Set(key string, value []byte, ttlMinutes uint64) {
	now := st.clock.Now()
	expire := now + ttlMinutes
	shard := st.shards[st.shardIndex(key)]
	shard.Set(key, value, expire)
}

func (st *Store) Get(key string) ([]byte, bool) {
	now := st.clock.Now()
	shard := st.shards[st.shardIndex(key)]
	return shard.Get(key, now)
}

func (st *Store) Delete(key string) {
	shard := st.shards[st.shardIndex(key)]
	shard.Delete(key)
}

// Janitor performs cleanup across all shards with proper synchronization.
func (st *Store) Janitor() {
	now := st.clock.Now()
	var wg sync.WaitGroup

	for _, shard := range st.shards {
		wg.Add(1)
		go func(s *Shard) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Janitor panic in shard: %v", r)
				}
			}()
			s.Janitor(now)
		}(shard)
	}

	wg.Wait()
}
