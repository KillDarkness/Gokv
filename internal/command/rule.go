package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
)

func registerRuleCommands(registry *Registry) {
	registry.Register(Command{Name: "RULE", Arity: -2, Handler: ruleCommand})
}

func ruleCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) < 2 {
		return protocol.Error("wrong number of arguments for 'rule' command")
	}

	switch strings.ToUpper(ctx.Args[1]) {
	case "SET":
		return ruleSetCommand(ctx)
	case "DEL":
		return ruleDelCommand(ctx)
	case "LIST":
		return ruleListCommand(ctx)
	default:
		return protocol.Error("unsupported RULE subcommand")
	}
}

func ruleSetCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) != 5 || !strings.EqualFold(ctx.Args[3], "ttl") {
		return protocol.Error("usage: RULE SET prefix ttl seconds")
	}
	seconds, err := strconv.ParseInt(ctx.Args[4], 10, 64)
	if err != nil || seconds <= 0 {
		return protocol.Error("ttl must be a positive integer")
	}
	ctx.Store.SetRule(ctx.Args[2], time.Duration(seconds)*time.Second)
	return protocol.SimpleString("OK")
}

func ruleDelCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) != 3 {
		return protocol.Error("usage: RULE DEL prefix")
	}
	if ctx.Store.DeleteRule(ctx.Args[2]) {
		return protocol.Integer(1)
	}
	return protocol.Integer(0)
}

func ruleListCommand(ctx *Context) protocol.Reply {
	if len(ctx.Args) != 2 {
		return protocol.Error("usage: RULE LIST")
	}
	rules := ctx.Store.Rules()
	replies := make(protocol.Array, 0, len(rules))
	for _, rule := range rules {
		replies = append(replies, protocol.Array{
			protocol.BulkString{Value: rule.Prefix},
			protocol.BulkString{Value: "ttl"},
			protocol.Integer(int64(rule.TTL / time.Second)),
		})
	}
	return replies
}
