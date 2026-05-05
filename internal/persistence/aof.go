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
	queue       chan []string
	async       bool
	lastErr     error
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
	args = append([]string(nil), args...)

	a.mu.Lock()
	queue := a.queue
	async := a.async
	lastErr := a.lastErr
	a.mu.Unlock()
	if lastErr != nil {
		return lastErr
	}
	if async {
		select {
		case queue <- args:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.writeCommandLocked(args); err != nil {
		return err
	}
	if a.FsyncPolicy == FsyncAlways {
		return a.file.Sync()
	}
	return nil
}

func (a *AOF) StartWriter(ctx context.Context, bufferSize int) <-chan struct{} {
	done := make(chan struct{})

	if !a.Enabled || a.FsyncPolicy == FsyncAlways {
		close(done)
		return done
	}
	if bufferSize <= 0 {
		bufferSize = 4096
	}

	a.mu.Lock()
	if a.queue == nil {
		a.queue = make(chan []string, bufferSize)
		a.async = true
	}
	queue := a.queue
	a.mu.Unlock()

	go func() {
		defer close(done)
		for {
			select {
			case args := <-queue:
				a.writeAsync(args)
			case <-ctx.Done():
				a.drainAsync(queue)
				_ = a.Sync()
				return
			}
		}
	}()

	return done
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
	return a.RewriteDatabases(ctx, []*store.Store{st})
}

func (a *AOF) RewriteDatabases(ctx context.Context, stores []*store.Store) error {
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

	writeErr := writeDatabasesAsAOF(file, stores)
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

func (a *AOF) writeAsync(args []string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.lastErr != nil {
		return
	}
	if err := a.writeCommandLocked(args); err != nil {
		a.lastErr = err
	}
}

func (a *AOF) drainAsync(queue chan []string) {
	for {
		select {
		case args := <-queue:
			a.writeAsync(args)
		default:
			return
		}
	}
}

func (a *AOF) writeCommandLocked(args []string) error {
	file, err := a.openAppendFile()
	if err != nil {
		return err
	}
	return protocol.WriteCommand(file, args)
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
	return writeDatabaseSnapshotAsAOF(file, 0, snapshot)
}

func writeDatabasesAsAOF(file *os.File, stores []*store.Store) error {
	for db, st := range stores {
		if err := writeDatabaseSnapshotAsAOF(file, db, st.Snapshot()); err != nil {
			return err
		}
	}
	return nil
}

func writeDatabaseSnapshotAsAOF(file *os.File, db int, snapshot map[string]store.SnapshotEntry) error {
	if len(snapshot) == 0 {
		return nil
	}
	if err := protocol.WriteCommand(file, []string{"SELECT", fmt.Sprintf("%d", db)}); err != nil {
		return err
	}
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
