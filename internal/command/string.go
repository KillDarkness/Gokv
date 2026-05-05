package command

import "github.com/KillDarkness/gokv/internal/protocol"

func registerStringCommands(registry *Registry) {
	registry.Register(Command{Name: "SET", Arity: 3, Handler: setCommand})
	registry.Register(Command{Name: "GET", Arity: 2, ReadOnly: true, Handler: getCommand})
	registry.Register(Command{Name: "INCR", Arity: 2, Handler: incrCommand})
	registry.Register(Command{Name: "DECR", Arity: 2, Handler: decrCommand})
}

func setCommand(ctx *Context) protocol.Reply {
	ctx.Store.Set(ctx.Args[1], ctx.Args[2])
	return protocol.SimpleString("OK")
}

func getCommand(ctx *Context) protocol.Reply {
	value, ok := ctx.Store.Get(ctx.Args[1])
	if !ok {
		return protocol.NullBulkString{}
	}
	return protocol.BulkString{Value: value}
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
