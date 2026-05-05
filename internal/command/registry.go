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
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
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

func (r *Registry) Dispatch(ctx context.Context, st *store.Store, args []string) protocol.Reply {
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

	return cmd.Handler(&Context{Context: ctx, Store: st, Args: args, StartedAt: time.Now()})
}

func validArity(arity int, got int) bool {
	if arity >= 0 {
		return got == arity
	}
	return got >= -arity
}
