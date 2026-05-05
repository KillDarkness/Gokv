# Usage

Gokv is a Redis-like in-memory key-value server. It speaks a basic RESP subset, so it can be used with `redis-cli` for supported commands.

## Run Locally

```sh
make run
```

Or run the binary entrypoint directly:

```sh
go run ./cmd/gokv
```

By default, Gokv listens on `0.0.0.0:6379`.

## Run With Docker

```sh
docker compose up -d --build
```

The Compose setup builds the local image, exposes port `6379` and mounts a Docker volume at `/data`.

Useful commands:

```sh
docker compose logs -f
docker compose restart gokv
docker compose down
```

## Test With redis-cli

```sh
redis-cli -p 6379 PING
redis-cli -p 6379 SET name kill
redis-cli -p 6379 GET name
redis-cli -p 6379 DEL name
```

## Multiple Databases

Gokv supports multiple logical databases when `GOKV_DATABASES` is greater than `1`.

```sh
GOKV_DATABASES=2 go run ./cmd/gokv
```

Select a database with `SELECT`:

```sh
redis-cli -p 6379 SELECT 1
redis-cli -p 6379 SET name db1
redis-cli -p 6379 GET name
```

With `redis-cli`, `-n` selects the database for a single command:

```sh
redis-cli -p 6379 -n 1 GET name
```

## Development Commands

```sh
make fmt
make vet
make test
make build
```
