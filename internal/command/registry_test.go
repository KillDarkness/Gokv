package command

import (
	"context"
	"strings"
	"testing"

	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/store"
)

func TestRegistryDispatchStringCommands(t *testing.T) {
	registry := NewDefaultRegistry()
	st := store.New()

	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"PING"}), "+PONG\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"SET", "name", "kill"}), "+OK\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"GET", "name"}), "$4\r\nkill\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"EXISTS", "name"}), ":1\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"TTL", "name"}), ":-1\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"EXPIRE", "name", "10"}), ":1\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"DEL", "name"}), ":1\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"GET", "name"}), "$-1\r\n")
	assertReply(t, registry.Dispatch(context.Background(), st, nil, []string{"TTL", "name"}), ":-2\r\n")
}

func assertReply(t *testing.T, reply protocol.Reply, want string) {
	t.Helper()

	var b strings.Builder
	if err := protocol.WriteReply(&b, reply); err != nil {
		t.Fatalf("WriteReply() error = %v", err)
	}
	if got := b.String(); got != want {
		t.Fatalf("reply = %q; want %q", got, want)
	}
}
