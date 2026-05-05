# Gokv

Gokv is a lightweight Redis-like in-memory key-value database written in Go.

Redis-compatible enough to be useful. Small enough to understand. Fast enough to be fun.

Current version: `0.2.0`

## Goal

Build a small, fast and maintainable in-memory key-value database using Go and the standard library.

## Run

```sh
make run
```

Or directly:

```sh
go run ./cmd/gokv
```

The server listens on `0.0.0.0:6379` by default.

## Test With redis-cli

```sh
redis-cli -p 6379 PING
redis-cli -p 6379 SET name kill
redis-cli -p 6379 GET name
redis-cli -p 6379 DEL name
```

## Docker

```sh
docker compose up -d --build
```

The Compose setup enables AOF persistence and stores data in the `gokv-data` Docker volume.

## Persistence

AOF persistence can be enabled with environment variables:

```sh
GOKV_APPENDONLY=true GOKV_AOF_PATH=data/appendonly.aof go run ./cmd/gokv
```

When AOF is enabled, write commands are appended to `appendonly.aof` and replayed on startup.

## Supported Commands

- `PING [message]`
- `SET key value`
- `GET key`
- `DEL key [key ...]`
- `EXISTS key [key ...]`

## Development

```sh
make fmt
make vet
make test
make build
```

## Roadmap

v0.1:

- TCP server
- RESP parser
- Command registry
- String commands
- Basic TTL

v0.2:

- AOF
- Snapshot
- Recovery

v0.3:

- List
- Hash
- Set

v0.4:

- Multi database
- Eviction
- Max memory

v0.5:

- Auth
- Go client
- Benchmarks
- Docker image
