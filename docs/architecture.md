# Architecture

Gokv is split by responsibility. The goal is to keep protocol parsing, command dispatching, storage and persistence independent enough to evolve without becoming hard to maintain.

## Packages

- `cmd/gokv`: binary entrypoint. Loads config, creates the app and handles process signals.
- `internal/app`: application lifecycle. Creates stores, registry, server, persistence and metrics.
- `internal/config`: default config and environment variable loading.
- `internal/server`: TCP listener, connection handling and selected database state per connection.
- `internal/protocol`: RESP parser, replies and writer helpers.
- `internal/command`: command registry, command metadata, arity validation and handlers.
- `internal/store`: in-memory data model, TTL, eviction, snapshots and thread safety.
- `internal/persistence`: AOF, AOF replay, AOF compaction and snapshots.
- `internal/metrics`: runtime counters exposed by `INFO`.
- `internal/log`: small logging wrapper.
- `pkg/client`: public client package placeholder.

## Request Flow

1. `internal/server` accepts a TCP connection.
2. `internal/protocol.Parser` reads a RESP command into string arguments.
3. `internal/server` handles connection-local commands such as `SELECT`.
4. `internal/command.Registry` normalizes the command name, validates arity and runs the handler.
5. Command handlers call `internal/store` and return `protocol.Reply` values.
6. Write commands are appended to AOF when persistence is enabled.
7. `internal/protocol.WriteReply` writes the RESP response.

## Store Model

Each logical database is a separate `store.Store` instance protected by `sync.RWMutex`.

The current supported value type is string. The data model already includes `ValueType` values for future list, hash and set support.

## Persistence Model

AOF records write commands and replays them at startup. On shutdown, Gokv compacts the AOF by writing the current keyspace back as `SET` and `EXPIRE` commands.

Snapshot persistence stores database `0` as JSON and is used only when AOF is disabled.
