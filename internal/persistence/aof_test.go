package persistence

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/store"
)

func TestAOFReplayRestoresWrittenCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")
	aof, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	ctx := context.Background()

	if err := aof.Append(ctx, []string{"SET", "name", "kill"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := aof.Append(ctx, []string{"SET", "lang", "go"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := aof.Append(ctx, []string{"DEL", "lang"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := aof.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	registry := command.NewDefaultRegistry(nil)
	st := store.New()
	replayAOF, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}

	if err := replayAOF.Replay(ctx, func(ctx context.Context, args []string) protocol.Reply {
		return registry.Dispatch(ctx, st, nil, args)
	}); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if got, ok := st.Get("name"); !ok || got != "kill" {
		t.Fatalf("Get(name) = %q, %v; want kill, true", got, ok)
	}
	if _, ok := st.Get("lang"); ok {
		t.Fatal("Get(lang) found deleted key after replay")
	}
}

func TestNewAOFRejectsInvalidFsyncPolicy(t *testing.T) {
	if _, err := NewAOF(true, filepath.Join(t.TempDir(), "appendonly.aof"), "sometimes"); err == nil {
		t.Fatal("NewAOF() error = nil; want error")
	}
}

func TestAOFRewriteCompactsCurrentState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")
	aof, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	ctx := context.Background()
	st := store.New()

	if err := aof.Append(ctx, []string{"SET", "name", "old"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := aof.Append(ctx, []string{"SET", "unused", "value"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	st.Set("name", "new")

	if err := aof.Rewrite(ctx, st); err != nil {
		t.Fatalf("Rewrite() error = %v", err)
	}

	registry := command.NewDefaultRegistry(nil)
	restored := store.New()
	replayAOF, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	if err := replayAOF.Replay(ctx, func(ctx context.Context, args []string) protocol.Reply {
		return registry.Dispatch(ctx, restored, nil, args)
	}); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if got, ok := restored.Get("name"); !ok || got != "new" {
		t.Fatalf("Get(name) = %q, %v; want new, true", got, ok)
	}
	if _, ok := restored.Get("unused"); ok {
		t.Fatal("Get(unused) found key removed by rewrite")
	}
}
