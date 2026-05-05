# Configuration

Gokv is configured with environment variables. Defaults are intentionally simple and safe for local development.

| Variable | Default | Description |
| --- | --- | --- |
| `GOKV_HOST` | `0.0.0.0` | TCP bind host. |
| `GOKV_PORT` | `6379` | TCP bind port. |
| `GOKV_DATABASES` | `1` | Number of logical databases. Must be at least `1`. |
| `GOKV_APPENDONLY` | `false` | Enables AOF persistence. |
| `GOKV_AOF_PATH` | `data/appendonly.aof` | Path to the AOF file. |
| `GOKV_AOF_FSYNC` | `always` | AOF fsync policy: `always`, `everysec` or `no`. |
| `GOKV_SNAPSHOT` | `false` | Enables snapshot persistence. |
| `GOKV_SNAPSHOT_PATH` | `data/dump.gokv` | Path to the snapshot file. |
| `GOKV_MAXKEYS` | `0` | Maximum number of keys per database. `0` means unlimited. |
| `GOKV_EVICTION` | `noeviction` | Eviction policy when `GOKV_MAXKEYS` is reached. |

## Example: AOF Persistence

```sh
GOKV_APPENDONLY=true \
GOKV_AOF_PATH=data/appendonly.aof \
GOKV_AOF_FSYNC=everysec \
go run ./cmd/gokv
```

## Example: Snapshot Persistence

```sh
GOKV_SNAPSHOT=true \
GOKV_SNAPSHOT_PATH=data/dump.gokv \
go run ./cmd/gokv
```

## Example: Multiple Databases

```sh
GOKV_DATABASES=16 go run ./cmd/gokv
```

## Example: Eviction

```sh
GOKV_MAXKEYS=1000 \
GOKV_EVICTION=allkeys-lru \
go run ./cmd/gokv
```

Supported eviction policies:

- `noeviction`
- `allkeys-random`
- `volatile-random`
- `allkeys-lru`
- `volatile-lru`
