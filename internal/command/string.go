package command

import (
	"strconv"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
)

func registerStringCommands(registry *Registry) {
	registry.Register(Command{Name: "SET", Arity: 3, Handler: setCommand})
	registry.Register(Command{Name: "GET", Arity: 2, ReadOnly: true, Handler: getCommand})
	registry.Register(Command{Name: "MSET", Arity: -3, Handler: msetCommand})
	registry.Register(Command{Name: "MGET", Arity: -2, ReadOnly: true, Handler: mgetCommand})
	registry.Register(Command{Name: "INCR", Arity: 2, Handler: incrCommand})
	registry.Register(Command{Name: "DECR", Arity: 2, Handler: decrCommand})
	registry.Register(Command{Name: "CAS", Arity: 4, Handler: casCommand})
	registry.Register(Command{Name: "SETNXEX", Arity: 4, Handler: setNXEXCommand})
	registry.Register(Command{Name: "GETSETEX", Arity: 4, Handler: getSetEXCommand})
}

func setCommand(ctx *Context) protocol.Reply {
	if err := ctx.Store.Set(ctx.Args[1], ctx.Args[2]); err != nil {
		return protocol.Error(err.Error())
	}
	return protocol.SimpleString("OK")
}

func getCommand(ctx *Context) protocol.Reply {
	value, ok := ctx.Store.Get(ctx.Args[1])
	if !ok {
		return protocol.NullBulkString{}
	}
	return protocol.BulkString{Value: value}
}

func msetCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args)%2 == 0 {
		return protocol.Error("wrong number of arguments for 'mset' command")
	}

	values := make(map[string]string, (len(ctx.Args)-1)/2)
	for i := 1; i < len(ctx.Args); i += 2 {
		values[ctx.Args[i]] = ctx.Args[i+1]
	}
	if err := ctx.Store.MSet(values); err != nil {
		return protocol.Error(err.Error())
	}
	return protocol.SimpleString("OK")
}

func mgetCommand(ctx *Context) protocol.Reply {
	results := ctx.Store.MGet(ctx.Args[1:]...)
	replies := make(protocol.Array, 0, len(results))
	for _, result := range results {
		if !result.OK {
			replies = append(replies, protocol.NullBulkString{})
			continue
		}
		replies = append(replies, protocol.BulkString{Value: result.Value})
	}
	return replies
}

func incrCommand(ctx *Context) protocol.Reply {
	return incrementCommand(ctx, 1)
}

func decrCommand(ctx *Context) protocol.Reply {
	return incrementCommand(ctx, -1)
}

func incrementCommand(ctx *Context, delta int64) protocol.Reply {
	value, err := ctx.Store.Increment(ctx.Args[1], delta)
	if err != nil {
		return protocol.Error("value is not an integer or out of range")
	}
	return protocol.Integer(value)
}

func casCommand(ctx *Context) protocol.Reply {
	swapped, err := ctx.Store.CompareAndSet(ctx.Args[1], ctx.Args[2], ctx.Args[3])
	if err != nil {
		return protocol.Error(err.Error())
	}
	if swapped {
		return protocol.Integer(1)
	}
	return protocol.Integer(0)
}

func setNXEXCommand(ctx *Context) protocol.Reply {
	ttl, ok := parsePositiveSeconds(ctx.Args[3])
	if !ok {
		return protocol.Error("ttl must be a positive integer")
	}
	set, err := ctx.Store.SetNXEX(ctx.Args[1], ctx.Args[2], ttl)
	if err != nil {
		return protocol.Error(err.Error())
	}
	if set {
		return protocol.Integer(1)
	}
	return protocol.Integer(0)
}

func getSetEXCommand(ctx *Context) protocol.Reply {
	ttl, ok := parsePositiveSeconds(ctx.Args[3])
	if !ok {
		return protocol.Error("ttl must be a positive integer")
	}
	oldValue, found, err := ctx.Store.GetSetEX(ctx.Args[1], ctx.Args[2], ttl)
	if err != nil {
		return protocol.Error(err.Error())
	}
	if !found {
		return protocol.NullBulkString{}
	}
	return protocol.BulkString{Value: oldValue}
}

func parsePositiveSeconds(value string) (time.Duration, bool) {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds <= 0 {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}
