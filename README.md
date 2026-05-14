# Cache Server

A high-performance, production-ready in-memory cache server written in Go with sharded storage, probabilistic eviction, and zero-syscall TTL tracking.

## Features

- ✅ **Sharded Architecture**: 256 shards with lock striping for high concurrency
- ✅ **SIEVE Eviction Policy**: Probabilistic O(1) eviction algorithm
- ✅ **Logical Clock**: Minute-granularity TTL without syscalls
- ✅ **Optional Compression**: Snappy compression for large values
- ✅ **Thread-Safe Metrics**: Atomic counters for monitoring
- ✅ **Graceful Shutdown**: Proper cleanup of all resources
- ✅ **RESTful API**: Simple HTTP interface

## Quick Start

### Prerequisites

- Go 1.23 or later

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd cache-server

# Build
go build -o cache-server ./cmd/server

# Run
export CACHE_MAX_ENTRIES=1000000
export CACHE_MAX_MEMORY=1073741824  # 1GB
export CACHE_DEFAULT_TTL_MIN=60
export CACHE_COMPRESSION=snappy
./cache-server
```

### Configuration

Configure via environment variables:

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CACHE_MAX_ENTRIES` | Yes* | Maximum number of cache entries | `1000000` |
| `CACHE_MAX_MEMORY` | Yes* | Maximum memory in bytes | `1073741824` |
| `CACHE_DEFAULT_TTL_MIN` | No | Default TTL in minutes (default: 60) | `60` |
| `CACHE_COMPRESSION` | No | Compression type: `snappy` or `none` | `snappy` |

*At least one of `CACHE_MAX_ENTRIES` or `CACHE_MAX_MEMORY` must be set.

### API Usage

#### Set a value

```bash
curl -X POST http://localhost:8080/api/v1/set \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:123",
    "value": {"name": "John", "age": 30},
    "ttl": 120
  }'
```

#### Get a value

```bash
curl http://localhost:8080/api/v1/get/user:123
```

#### Delete a value

```bash
curl -X DELETE http://localhost:8080/api/v1/del/user:123
```

#### Get metrics

```bash
curl http://localhost:8080/api/v1/metrics
```

#### Health check

```bash
curl http://localhost:8080/health
```

## Architecture

### Sharding

The cache uses 256 shards to minimize lock contention. Keys are distributed using xxHash for uniform distribution.

### Eviction Policy

Implements the SIEVE algorithm - a probabilistic eviction policy that:
- Operates in O(1) amortized time
- Evicts entries randomly with probability based on memory pressure
- Each shard has its own evictor instance to avoid contention

### TTL Management

Uses a logical clock that increments every minute, avoiding expensive syscalls on every operation. Expired entries are removed:
1. **Lazily** on Get (immediate deletion)
2. **Proactively** by background janitor (runs every minute)

### Compression

When enabled, values ≥256 bytes are compressed using Snappy. Smaller values skip compression to avoid overhead.

## Performance Characteristics

- **Set**: O(1) average, O(n) worst case during eviction
- **Get**: O(1) with lazy deletion
- **Delete**: O(1)
- **Memory**: Sharded to reduce lock contention
- **Concurrency**: Read-write locks per shard

## Monitoring

Metrics available at `/api/v1/metrics`:

```json
{
  "sets": 1000,
  "hits": 850,
  "misses": 150,
  "deletes": 50,
  "corruptions": 0
}
```

## Production Considerations

### Resource Limits

- **Connection Limits**: 
  - Read timeout: 10s
  - Write timeout: 10s
  - Idle timeout: 120s
  - Max body size: 10MB

### Graceful Shutdown

The server handles SIGINT and SIGTERM signals:
1. Stops accepting new connections
2. Cancels background janitor
3. Stops logical clock
4. Waits up to 10s for in-flight requests
5. Exits cleanly

### Memory Management

Memory limits are divided across shards. Each shard independently enforces its portion of the limit.

## Development

### Running Tests

```bash
go test ./... -v -race
```

### Benchmarks

```bash
go test ./... -bench=. -benchmem
```

### Code Quality

```bash
# Vet
go vet ./...

# Format
go fmt ./...

# Race detector
go run -race ./cmd/server
```

## Known Limitations

1. **No Persistence**: Data is lost on restart (in-memory only)
2. **Single Node**: No clustering or replication
3. **Minute Granularity**: TTL precision is 1 minute
4. **No Authentication**: Add reverse proxy for auth/TLS

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]
