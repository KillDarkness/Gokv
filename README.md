# Gokv

Gokv is a lightweight Redis-like in-memory key-value database written in Go.

Redis-compatible enough to be useful. Small enough to understand. Fast enough to be fun.

Current version: `0.2.10`

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
GOKV_APPENDONLY=true GOKV_AOF_PATH=data/appendonly.aof GOKV_AOF_FSYNC=everysec go run ./cmd/gokv
```

When AOF is enabled, write commands are appended to `appendonly.aof` and replayed on startup.
Supported fsync policies are `always`, `everysec` and `no`.
On shutdown, Gokv rewrites the AOF with the current keyspace to compact old writes.

Snapshot persistence can be enabled with:

```sh
GOKV_SNAPSHOT=true GOKV_SNAPSHOT_PATH=data/dump.gokv go run ./cmd/gokv
```

When both AOF and snapshot are enabled, AOF is used as the recovery source.

## Supported Commands

- `PING [message]`
- `SET key value`
- `GET key`
- `DEL key [key ...]`
- `EXISTS key [key ...]`
- `EXPIRE key seconds`
- `TTL key`
- `INCR key`
- `DECR key`
- `MSET key value [key value ...]`
- `MGET key [key ...]`
- `FLUSHDB`
- `INFO`
- `SELECT index`
- `RULE SET prefix ttl seconds`
- `RULE DEL prefix`
- `RULE LIST`

`INFO` exposes basic server, client, command and keyspace metrics.

Eviction can be enabled with `GOKV_MAXKEYS` and `GOKV_EVICTION`. Supported policies are `noeviction`, `allkeys-random`, `volatile-random`, `allkeys-lru` and `volatile-lru`.
Multiple logical databases can be enabled with `GOKV_DATABASES` and selected with `SELECT`.
Rules can apply automatic TTLs by prefix, for example `RULE SET session: ttl 1800` makes future `session:*` writes expire automatically.

## Development

```sh
make fmt
make vet
make test
make build
```

## Documentation

Full documentation is available in [`docs/`](docs/README.md):

- [Usage](docs/usage.md)
- [Configuration](docs/configuration.md)
- [Commands](docs/commands.md)
- [Persistence](docs/persistence.md)
- [Protocol](docs/protocol.md)
- [Architecture](docs/architecture.md)
- [Project Standards](docs/project-standards.md)
- [Best Practices](docs/best-practices.md)

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
