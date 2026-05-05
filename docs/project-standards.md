# Project Standards

These standards keep Gokv small, predictable and easy to maintain.

## Go Style

- Use idiomatic Go and the standard library first.
- Keep functions small enough to understand without jumping across many files.
- Return errors explicitly. Do not use `panic` for normal control flow.
- Prefer concrete types until an interface has a real consumer.
- Run `gofmt`, `go vet` and `go test ./...` before committing.

## Package Boundaries

- Protocol parsing belongs in `internal/protocol`.
- TCP connection handling belongs in `internal/server`.
- Command dispatch and command handlers belong in `internal/command`.
- Data mutation and TTL behavior belong in `internal/store`.
- AOF and snapshot code belongs in `internal/persistence`.
- Config loading belongs in `internal/config`.

Do not mix protocol parsing, command dispatch and store internals in the same package.

## Command Handlers

Command handlers should:

- Validate command-specific arguments.
- Use `store.Store` methods instead of touching map internals.
- Return `protocol.Reply` values.
- Avoid writing directly to the network.
- Avoid persistence-specific code; the registry handles AOF append after successful write commands.

## Tests

Tests should live close to the package being tested.

Minimum expectations:

- Store behavior has unit tests.
- RESP parser and writer behavior has unit tests.
- Command registry and command handlers have tests.
- Persistence replay/rewrite behavior has tests.

## Commits

Use Conventional Commits in English:

- `feat: add ttl commands`
- `fix: handle expired keys during replay`
- `docs: add command examples`
- `test: cover aof rewrite`
- `chore: update docker setup`

Prefer one commit per functional change.
