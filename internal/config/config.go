// Package config loads server configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"cache-server/pkg/types"
)

// Load reads configuration from environment variables with validation.
func Load() (*types.Config, error) {
	maxEntriesStr := os.Getenv("CACHE_MAX_ENTRIES")
	maxMemoryStr := os.Getenv("CACHE_MAX_MEMORY")

	defaultTTLStr := os.Getenv("CACHE_DEFAULT_TTL_MIN")
	comp := os.Getenv("CACHE_COMPRESSION")
	fmt.Println("maxEntires :", maxEntriesStr, "maxMemory :", maxMemoryStr, "ttl :", defaultTTLStr, "compresion :", comp)

	var maxEntries, maxMemory int64
	var defaultTTL uint64
	var err error

	// Parse maxEntries if provided
	if maxEntriesStr != "" {
		maxEntries, err = strconv.ParseInt(maxEntriesStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_MAX_ENTRIES: %w", err)
		}
		if maxEntries < 0 {
			return nil, errors.New("CACHE_MAX_ENTRIES must be non-negative")
		}
	}

	// Parse maxMemory if provided
	if maxMemoryStr != "" {
		maxMemory, err = strconv.ParseInt(maxMemoryStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_MAX_MEMORY: %w", err)
		}
		if maxMemory < 0 {
			return nil, errors.New("CACHE_MAX_MEMORY must be non-negative")
		}
	}

	// Require at least one limit
	if maxEntries == 0 && maxMemory == 0 {
		return nil, errors.New("must set at least one of CACHE_MAX_ENTRIES or CACHE_MAX_MEMORY")
	}

	// Parse defaultTTL if provided
	if defaultTTLStr != "" {
		defaultTTL, err = strconv.ParseUint(defaultTTLStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_DEFAULT_TTL_MIN: %w", err)
		}
	} else {
		defaultTTL = 60 // Default to 60 minutes
	}

	// Validate compression type
	if comp != "" && comp != "snappy" && comp != "none" {
		return nil, fmt.Errorf("invalid CACHE_COMPRESSION: must be 'snappy' or 'none', got '%s'", comp)
	}

	return &types.Config{
		MaxEntries:  maxEntries,
		MaxMemory:   maxMemory,
		DefaultTTL:  defaultTTL,
		Compression: comp,
	}, nil
}
