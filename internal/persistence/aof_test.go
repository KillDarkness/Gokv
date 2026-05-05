package persistence

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
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
	if err := st.Set("name", "new"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

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
		if len(args) == 2 && strings.EqualFold(args[0], "SELECT") {
			return protocol.SimpleString("OK")
		}
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

func TestAOFRewritePersistsMultipleDatabases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")
	aof, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	ctx := context.Background()
	dbs := []*store.Store{store.New(), store.New()}
	if err := dbs[0].Set("name", "db0"); err != nil {
		t.Fatalf("Set(db0) error = %v", err)
	}
	if err := dbs[1].Set("name", "db1"); err != nil {
		t.Fatalf("Set(db1) error = %v", err)
	}

	if err := aof.RewriteDatabases(ctx, dbs); err != nil {
		t.Fatalf("RewriteDatabases() error = %v", err)
	}

	registry := command.NewDefaultRegistry(nil)
	restored := []*store.Store{store.New(), store.New()}
	selectedDB := 0
	replayAOF, err := NewAOF(true, path, string(FsyncAlways))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	if err := replayAOF.Replay(ctx, func(ctx context.Context, args []string) protocol.Reply {
		if len(args) == 2 && strings.EqualFold(args[0], "SELECT") {
			db, err := strconv.Atoi(args[1])
			if err != nil || db < 0 || db >= len(restored) {
				return protocol.Error("DB index is out of range")
			}
			selectedDB = db
			return protocol.SimpleString("OK")
		}
		return registry.Dispatch(ctx, restored[selectedDB], nil, args)
	}); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if got, ok := restored[0].Get("name"); !ok || got != "db0" {
		t.Fatalf("db0 Get(name) = %q, %v; want db0, true", got, ok)
	}
	if got, ok := restored[1].Get("name"); !ok || got != "db1" {
		t.Fatalf("db1 Get(name) = %q, %v; want db1, true", got, ok)
	}
}
