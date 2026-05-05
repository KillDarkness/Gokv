package command

import (
	"strconv"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
)

func registerKeyCommands(registry *Registry) {
	registry.Register(Command{Name: "DEL", Arity: -2, Handler: delCommand})
	registry.Register(Command{Name: "EXISTS", Arity: -2, ReadOnly: true, Handler: existsCommand})
	registry.Register(Command{Name: "EXPIRE", Arity: 3, Handler: expireCommand})
	registry.Register(Command{Name: "TTL", Arity: 2, ReadOnly: true, Handler: ttlCommand})
	registry.Register(Command{Name: "FLUSHDB", Arity: 1, Handler: flushDBCommand})
}

func delCommand(ctx *Context) protocol.Reply {
	return protocol.Integer(ctx.Store.Delete(ctx.Args[1:]...))
}

func existsCommand(ctx *Context) protocol.Reply {
	return protocol.Integer(ctx.Store.Exists(ctx.Args[1:]...))
}

func expireCommand(ctx *Context) protocol.Reply {
	seconds, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return protocol.Error("value is not an integer or out of range")
	}
	if ctx.Store.Expire(ctx.Args[1], time.Duration(seconds)*time.Second) {
		return protocol.Integer(1)
	}
	return protocol.Integer(0)
}

func ttlCommand(ctx *Context) protocol.Reply {
	ttl, exists, hasTTL := ctx.Store.TTL(ctx.Args[1])
	if !exists {
		return protocol.Integer(-2)
	}
	if !hasTTL {
		return protocol.Integer(-1)
	}
	return protocol.Integer(int64(ttl / time.Second))
}

func flushDBCommand(ctx *Context) protocol.Reply {
	ctx.Store.FlushDB()
	return protocol.SimpleString("OK")
}
