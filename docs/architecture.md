# Architecture

Gokv is split by responsibility:

- `cmd/gokv` starts the application.
- `internal/app` owns application lifecycle.
- `internal/server` accepts TCP connections.
- `internal/protocol` parses and writes RESP.
- `internal/command` dispatches commands.
- `internal/store` stores in-memory data.
- `internal/persistence` keeps placeholders for AOF, snapshot and replay.
- `internal/config` defines runtime defaults.

The first version keeps one in-memory database protected by `sync.RWMutex` and supports string values.
