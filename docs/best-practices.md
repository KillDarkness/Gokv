# Best Practices

This guide explains how to add features without making the codebase harder to maintain.

## Adding A Command

1. Add or reuse a store method in `internal/store`.
2. Add the command handler in `internal/command`.
3. Register the command in the correct registration function.
4. Set `ReadOnly` correctly so AOF persistence only records writes.
5. Add command tests in `internal/command`.
6. Add store tests when behavior changes storage semantics.
7. Update `docs/commands.md` and `README.md` if user-facing behavior changes.

## Adding Store Behavior

Keep store operations thread-safe and explicit.

- Hold the mutex for the shortest practical scope.
- Clean expired keys lazily when reading or checking existence.
- Preserve TTL when mutating an existing key only when the command semantics require it.
- Avoid exposing the internal map directly.

## Adding Persistence Behavior

Persistence should replay into the command layer when possible so behavior stays consistent.

- AOF should record accepted write commands only.
- Replay should not append replayed commands back into AOF.
- Rewrite/compaction should produce a minimal recoverable command stream.
- Snapshot should avoid storing expired keys.

## Adding Config

Add config in this order:

1. Field in `config.Config`.
2. Default value in `config.Default`.
3. Environment variable parsing in `config.Load`.
4. Documentation in `docs/configuration.md`.
5. Runtime wiring in `internal/app` or the relevant package.

## Compatibility

Gokv should be Redis-compatible enough to be useful, not a full Redis clone.

When behavior differs from Redis, document it clearly.

## Performance

Prefer simple implementations first. Optimize once behavior is correct and measured.

Good defaults:

- Avoid unnecessary allocations in hot paths.
- Keep command dispatch straightforward.
- Use `sync.RWMutex` or `sync.Mutex` intentionally.
- Do not introduce external dependencies without a clear benefit.
