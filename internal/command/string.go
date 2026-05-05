package command

import "github.com/KillDarkness/gokv/internal/protocol"

func registerStringCommands(registry *Registry) {
	registry.Register(Command{Name: "SET", Arity: 3, Handler: setCommand})
	registry.Register(Command{Name: "GET", Arity: 2, ReadOnly: true, Handler: getCommand})
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
