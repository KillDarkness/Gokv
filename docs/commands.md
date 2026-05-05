# Commands

This document describes the commands currently supported by Gokv and shows examples with `redis-cli`.

## PING

Checks whether the server is reachable.

```sh
redis-cli -p 6379 PING
# PONG

redis-cli -p 6379 PING hello
# hello
```

## SET

Stores a string value.

```sh
redis-cli -p 6379 SET name kill
# OK
```

## GET

Returns a string value or null when the key does not exist.

```sh
redis-cli -p 6379 GET name
# kill

redis-cli -p 6379 GET missing
# (nil)
```

## DEL

Deletes one or more keys and returns the number of removed keys.

```sh
redis-cli -p 6379 DEL name
# 1
```

## EXISTS

Returns how many keys exist.

```sh
redis-cli -p 6379 EXISTS name missing
# 1
```

## EXPIRE

Sets a TTL in seconds. Returns `1` when the key exists and `0` otherwise.

```sh
redis-cli -p 6379 SET session active
redis-cli -p 6379 EXPIRE session 10
# 1
```

## TTL

Returns the remaining TTL in seconds.

```sh
redis-cli -p 6379 TTL session
# 9
```

Return values:

- `-2`: key does not exist.
- `-1`: key exists but has no expiration.
- `>= 0`: remaining TTL in seconds.

## INCR

Increments an integer string value by `1`. Missing keys start at `0` before incrementing.

```sh
redis-cli -p 6379 INCR counter
# 1
```

## DECR

Decrements an integer string value by `1`. Missing keys start at `0` before decrementing.

```sh
redis-cli -p 6379 DECR counter
# 0
```

## CAS

Compares the current value and swaps it atomically when it matches. Returns `1` when swapped and `0` otherwise.

```sh
redis-cli -p 6379 SET config:v 1
redis-cli -p 6379 CAS config:v 1 2
# 1

redis-cli -p 6379 CAS config:v 1 3
# 0
```

## SETNXEX

Sets a value with a TTL only when the key does not exist. Returns `1` when set and `0` otherwise.

```sh
redis-cli -p 6379 SETNXEX lock:job token 30
# 1
```

## GETSETEX

Atomically returns the previous value and stores a new value with a TTL.

```sh
redis-cli -p 6379 GETSETEX session:abc new-token 1800
# token
```

## MSET

Stores multiple key/value pairs.

```sh
redis-cli -p 6379 MSET a 1 b 2
# OK
```

## MGET

Returns multiple values in order.

```sh
redis-cli -p 6379 MGET a missing b
# 1
# (nil)
# 2
```

## FLUSHDB

Removes all keys from the selected database.

```sh
redis-cli -p 6379 FLUSHDB
# OK
```

## INFO

Returns server, client, command and keyspace metrics.

```sh
redis-cli -p 6379 INFO
```

Example sections:

- `# Server`
- `# Clients`
- `# Stats`
- `# Keyspace`

## SELECT

Selects a logical database for the current connection.

```sh
redis-cli -p 6379 SELECT 1
# OK
```

For one-shot `redis-cli` commands, use `-n`:

```sh
redis-cli -p 6379 -n 1 SET name db1
redis-cli -p 6379 -n 1 GET name
```

## RULE

Manages automatic key rules for the selected database. The first supported rule type is prefix TTL.

```sh
redis-cli -p 6379 RULE SET session: ttl 1800
# OK

redis-cli -p 6379 SET session:abc token
# OK

redis-cli -p 6379 TTL session:abc
# 1800
```

List rules:

```sh
redis-cli -p 6379 RULE LIST
```

Delete a rule:

```sh
redis-cli -p 6379 RULE DEL session:
# 1
```

When multiple rules match, the longest prefix wins.
