package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/KillDarkness/gokv/internal/store"
)

type Snapshot struct {
	Enabled bool
	Path    string
}

func NewSnapshot(enabled bool, path string) *Snapshot {
	if path == "" {
		path = "data/dump.gokv"
	}
	return &Snapshot{Enabled: enabled, Path: path}
}

func (s *Snapshot) Save(ctx context.Context, st *store.Store) error {
	if !s.Enabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}

	tmpPath := s.Path + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	encodeErr := json.NewEncoder(file).Encode(st.Snapshot())
	syncErr := file.Sync()
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmpPath, s.Path)
}

func (s *Snapshot) Load(ctx context.Context, st *store.Store) error {
	if !s.Enabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := os.Open(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	snapshot := make(map[string]store.SnapshotEntry)
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return err
	}
	st.Restore(snapshot)
	return nil
}
