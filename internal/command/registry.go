package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/store"
)

type Handler func(ctx *Context) protocol.Reply

type Command struct {
	Name     string
	Arity    int
	ReadOnly bool
	Handler  Handler
}

type Registry struct {
	commands  map[string]Command
	startedAt time.Time
}

func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command), startedAt: time.Now()}
}

func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registerServerCommands(registry)
	registerStringCommands(registry)
	registerKeyCommands(registry)
	return registry
}

func (r *Registry) Register(cmd Command) {
	cmd.Name = strings.ToUpper(cmd.Name)
	r.commands[cmd.Name] = cmd
}

func (r *Registry) Dispatch(ctx context.Context, st *store.Store, appender Appender, args []string) protocol.Reply {
	if len(args) == 0 {
		return protocol.Error("empty command")
	}

	name := strings.ToUpper(args[0])
	cmd, ok := r.commands[name]
	if !ok {
		return protocol.Error(fmt.Sprintf("unknown command '%s'", args[0]))
	}
	if !validArity(cmd.Arity, len(args)) {
		return protocol.Error(fmt.Sprintf("wrong number of arguments for '%s' command", strings.ToLower(cmd.Name)))
	}

	commandCtx := &Context{Context: ctx, Store: st, Appender: appender, Args: args, StartedAt: r.startedAt}
	reply := cmd.Handler(commandCtx)
	if cmd.ReadOnly || appender == nil {
		return reply
	}
	if _, ok := reply.(protocol.Error); ok {
		return reply
	}
	if err := appender.Append(ctx, args); err != nil {
		return protocol.Error(fmt.Sprintf("could not persist command: %v", err))
	}
	return reply
}

func validArity(arity int, got int) bool {
	if arity >= 0 {
		return got == arity
	}
	return got >= -arity
}
