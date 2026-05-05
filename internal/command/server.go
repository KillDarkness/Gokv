package command

import (
	"fmt"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/version"
)

func registerServerCommands(registry *Registry) {
	registry.Register(Command{Name: "PING", Arity: -1, ReadOnly: true, Handler: pingCommand})
	registry.Register(Command{Name: "INFO", Arity: 1, ReadOnly: true, Handler: infoCommand})
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

func infoCommand(ctx *Context) protocol.Reply {
	uptime := int64(time.Since(ctx.StartedAt).Seconds())
	info := fmt.Sprintf("# Server\r\ngokv_version:%s\r\nuptime_in_seconds:%d\r\n\r\n# Keyspace\r\ndb0:keys=%d\r\n", version.Version, uptime, ctx.Store.Size())
	return protocol.BulkString{Value: info}
}
