package persistence

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/KillDarkness/gokv/internal/protocol"
)

type AOF struct {
	Enabled bool
	Path    string
	mu      sync.Mutex
	file    *os.File
}

func NewAOF(enabled bool, path string) *AOF {
	if path == "" {
		path = "data/appendonly.aof"
	}
	return &AOF{Enabled: enabled, Path: path}
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
	return file.Sync()
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
