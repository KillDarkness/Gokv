package command

import "github.com/KillDarkness/gokv/internal/protocol"

func registerKeyCommands(registry *Registry) {
	registry.Register(Command{Name: "DEL", Arity: -2, Handler: delCommand})
	registry.Register(Command{Name: "EXISTS", Arity: -2, ReadOnly: true, Handler: existsCommand})
}

func delCommand(ctx *Context) protocol.Reply {
	return protocol.Integer(ctx.Store.Delete(ctx.Args[1:]...))
}

func existsCommand(ctx *Context) protocol.Reply {
	return protocol.Integer(ctx.Store.Exists(ctx.Args[1:]...))
}
