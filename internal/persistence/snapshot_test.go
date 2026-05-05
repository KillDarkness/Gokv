package persistence

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/KillDarkness/gokv/internal/store"
)

func TestSnapshotSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "dump.gokv")
	snapshot := NewSnapshot(true, path)
	st := store.New()
	st.Set("name", "kill")

	if err := snapshot.Save(ctx, st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	restored := store.New()
	if err := snapshot.Load(ctx, restored); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, ok := restored.Get("name"); !ok || got != "kill" {
		t.Fatalf("Get() = %q, %v; want kill, true", got, ok)
	}
}
