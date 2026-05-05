package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/store"
)

type FsyncPolicy string

const (
	FsyncAlways   FsyncPolicy = "always"
	FsyncEverySec FsyncPolicy = "everysec"
	FsyncNo       FsyncPolicy = "no"
)

type AOF struct {
	Enabled     bool
	Path        string
	FsyncPolicy FsyncPolicy
	mu          sync.Mutex
	file        *os.File
}

func NewAOF(enabled bool, path string, fsync string) (*AOF, error) {
	if path == "" {
		path = "data/appendonly.aof"
	}
	policy := FsyncPolicy(fsync)
	if policy == "" {
		policy = FsyncAlways
	}
	if !validFsyncPolicy(policy) {
		return nil, fmt.Errorf("invalid AOF fsync policy %q", fsync)
	}
	return &AOF{Enabled: enabled, Path: path, FsyncPolicy: policy}, nil
}

func (a *AOF) Append(ctx context.Context, args []string) error {
	if !a.Enabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := a.openAppendFile()
	if err != nil {
		return err
	}
	if err := protocol.WriteCommand(file, args); err != nil {
		return err
	}
	if a.FsyncPolicy == FsyncAlways {
		return file.Sync()
	}
	return nil
}

func (a *AOF) StartSyncer(ctx context.Context, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		if !a.Enabled || a.FsyncPolicy != FsyncEverySec {
			<-ctx.Done()
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				_ = a.Sync()
				return
			case <-ticker.C:
				_ = a.Sync()
			}
		}
	}()

	return done
}

func (a *AOF) Sync() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		return nil
	}
	return a.file.Sync()
}

func (a *AOF) Rewrite(ctx context.Context, st *store.Store) error {
	if !a.Enabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		if err := a.file.Close(); err != nil {
			return err
		}
		a.file = nil
	}
	if err := os.MkdirAll(filepath.Dir(a.Path), 0o755); err != nil {
		return err
	}

	tmpPath := a.Path + ".rewrite"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	writeErr := writeSnapshotAsAOF(file, st.Snapshot())
	syncErr := file.Sync()
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmpPath, a.Path)
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		return nil
	}
	err := a.file.Close()
	a.file = nil
	return err
}

func (a *AOF) openAppendFile() (*os.File, error) {
	if a.file != nil {
		return a.file, nil
	}
	if err := os.MkdirAll(filepath.Dir(a.Path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(a.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	a.file = file
	return file, nil
}

func validFsyncPolicy(policy FsyncPolicy) bool {
	switch policy {
	case FsyncAlways, FsyncEverySec, FsyncNo:
		return true
	default:
		return false
	}
}

func writeSnapshotAsAOF(file *os.File, snapshot map[string]store.SnapshotEntry) error {
	now := time.Now().UnixNano()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := snapshot[key]
		if entry.ExpiresAt > 0 && entry.ExpiresAt <= now {
			continue
		}
		if err := protocol.WriteCommand(file, []string{"SET", key, entry.Value}); err != nil {
			return err
		}
		if entry.ExpiresAt > 0 {
			remaining := entry.ExpiresAt - now
			seconds := (remaining + int64(time.Second) - 1) / int64(time.Second)
			if err := protocol.WriteCommand(file, []string{"EXPIRE", key, fmt.Sprintf("%d", seconds)}); err != nil {
				return err
			}
		}
	}
	return nil
}
