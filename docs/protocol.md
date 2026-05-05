# Protocol

Gokv uses a small RESP subset compatible with supported `redis-cli` commands.

Example request:

```text
*2\r\n$4\r\nPING\r\n$4\r\ntest\r\n
```

Example response:

```text
$4\r\ntest\r\n
```

Implemented reply types:

- Simple strings
- Errors
- Integers
- Bulk strings
- Null bulk strings
- Arrays

## Command Format

Commands are expected as RESP arrays of bulk strings:

```text
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nkill\r\n
```

Inline commands are also accepted for simple manual tests:

```text
PING\r\n
```

## Compatibility Scope

Gokv is not a full Redis server. It is compatible with Redis clients only for commands and RESP types implemented by this project.
