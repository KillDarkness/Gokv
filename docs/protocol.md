# Protocol

Gokv uses a small RESP subset compatible with basic `redis-cli` commands.

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
