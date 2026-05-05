package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/version"
)

func registerServerCommands(registry *Registry) {
	registry.Register(Command{Name: "PING", Arity: -1, ReadOnly: true, Handler: pingCommand})
	registry.Register(Command{Name: "INFO", Arity: 1, ReadOnly: true, Handler: infoCommand})
	registry.Register(Command{Name: "CONFIG", Arity: -2, ReadOnly: true, Handler: configCommand})
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
	info := fmt.Sprintf("# Server\r\ngokv_version:%s\r\nuptime_in_seconds:%d\r\n\r\n# Clients\r\nconnected_clients:%d\r\ntotal_connections_received:%d\r\n\r\n# Stats\r\ntotal_commands_processed:%d\r\ntotal_errors:%d\r\n\r\n# Keyspace\r\ndb0:keys=%d\r\n", version.Version, uptime, ctx.Metrics.ActiveConnections(), ctx.Metrics.TotalConnections(), ctx.Metrics.Commands(), ctx.Metrics.Errors(), ctx.Store.Size())
	return protocol.BulkString{Value: info}
}

func configCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) >= 2 && strings.EqualFold(ctx.Args[1], "GET") {
		if len(ctx.Args) == 2 || ctx.Args[2] == "*" {
			return protocol.Array{
				protocol.BulkString{Value: "save"},
				protocol.BulkString{Value: ""},
				protocol.BulkString{Value: "appendonly"},
				protocol.BulkString{Value: "no"},
			}
		}
		switch strings.ToLower(ctx.Args[2]) {
		case "save":
			return protocol.Array{protocol.BulkString{Value: "save"}, protocol.BulkString{Value: ""}}
		case "appendonly":
			return protocol.Array{protocol.BulkString{Value: "appendonly"}, protocol.BulkString{Value: "no"}}
		default:
			return protocol.Array{}
		}
	}
	return protocol.Error("unsupported CONFIG subcommand")
}
