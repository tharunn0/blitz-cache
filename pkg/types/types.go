// Package types defines shared types for the cache server.
package types

type Entry struct {
	Value     []byte
	ExpireTick uint64
}

type Config struct {
	MaxEntries int64
	MaxMemory  int64 // bytes
	DefaultTTL uint64 // minutes
	Compression string // \"gzip\" or \"snappy\"
}