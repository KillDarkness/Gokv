package command

import "github.com/KillDarkness/gokv/internal/protocol"

func registerServerCommands(registry *Registry) {
	registry.Register(Command{Name: "PING", Arity: -1, ReadOnly: true, Handler: pingCommand})
}

func pingCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) == 1 {
		return protocol.SimpleString("PONG")
	}
	if len(ctx.Args) == 2 {
		return protocol.BulkString{Value: ctx.Args[1]}
	}
	return protocol.Error("wrong number of arguments for 'ping' command")
}
