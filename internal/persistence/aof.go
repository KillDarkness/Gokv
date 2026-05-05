package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/KillDarkness/gokv/internal/protocol"
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
