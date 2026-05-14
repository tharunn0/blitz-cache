// Package service provides the cache service layer.
package service

import (
	"context"
	"time"

	"cache-server/internal/cache/store/sharded"
	"cache-server/internal/clock"
	"cache-server/internal/metrics"
	"cache-server/pkg/types"

	"github.com/klauspost/compress/snappy"
)

const compressionThreshold = 256 // bytes - only compress values larger than this

type Service struct {
	store   *sharded.Store
	clock   *clock.Clock
	metrics *metrics.CacheMetrics
	config  *types.Config
}

func NewService(cfg *types.Config) *Service {
	clk := clock.NewClock()
	// Create separate evictor per shard by passing limits to store
	store := sharded.NewStore(clk, cfg.MaxEntries, cfg.MaxMemory)
	m := &metrics.CacheMetrics{}

	s := &Service{
		store:   store,
		clock:   clk,
		metrics: m,
		config:  cfg,
	}
	return s
}

func (s *Service) Set(key string, value []byte, ttl uint64) error {
	if ttl == 0 {
		ttl = s.config.DefaultTTL
	}

	// Compress if enabled and value is large enough
	if s.config.Compression == "snappy" && len(value) >= compressionThreshold {
		value = snappy.Encode(nil, value)
	}

	s.store.Set(key, value, ttl)
	s.metrics.IncSet()
	return nil
}

func (s *Service) Get(key string) ([]byte, bool) {
	value, found := s.store.Get(key)
	if !found {
		s.metrics.IncMiss()
		return nil, false
	}
	s.metrics.IncHit()

	// Decompress if needed
	if s.config.Compression == "snappy" && len(value) > 0 {
		var err error
		value, err = snappy.Decode(nil, value)
		if err != nil {
			s.metrics.IncCorruption()
			return nil, false
		}
	}

	return value, true
}

func (s *Service) Delete(key string) {
	s.store.Delete(key)
	s.metrics.IncDel()
}

func (s *Service) Janitor(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.store.Janitor()
		}
	}
}

func (s *Service) Metrics() *metrics.CacheMetrics {
	return s.metrics.Get()
}

// Stop gracefully stops the service and cleans up resources.
func (s *Service) Stop() {
	s.clock.Stop()
}
