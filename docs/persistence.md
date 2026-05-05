# Persistence

Gokv is in-memory by default. Persistence is optional and can be enabled with AOF or snapshot configuration.

## AOF

AOF stores accepted write commands in RESP format. On startup, Gokv replays the AOF to rebuild memory state.

Enable AOF:

```sh
GOKV_APPENDONLY=true go run ./cmd/gokv
```

Configure the path:

```sh
GOKV_APPENDONLY=true GOKV_AOF_PATH=data/appendonly.aof go run ./cmd/gokv
```

Configure fsync:

```sh
GOKV_APPENDONLY=true GOKV_AOF_FSYNC=everysec go run ./cmd/gokv
```

Fsync policies:

- `always`: sync after each write. Safest, slowest.
- `everysec`: sync periodically. Balanced default for Docker Compose.
- `no`: rely on OS flushing. Fastest, least durable.

## AOF Compaction

On shutdown, Gokv rewrites the AOF using the current keyspace. This removes old overwritten writes and deleted keys.

For multiple databases, the rewritten AOF includes `SELECT` commands so replay restores each database correctly.

## Snapshot

Snapshot persistence stores the current database state as JSON. It is intended as a simple recovery mechanism.

Enable snapshot:

```sh
GOKV_SNAPSHOT=true GOKV_SNAPSHOT_PATH=data/dump.gokv go run ./cmd/gokv
```

Snapshot currently saves and loads database `0`.

## Recovery Priority

If both AOF and snapshot are enabled, AOF is used as the recovery source.

## Docker Persistence

The default Compose file enables AOF and stores data in the `gokv-data` volume.

```sh
docker compose up -d --build
docker compose restart gokv
```

Data survives container restarts and recreations as long as the Docker volume is kept.
