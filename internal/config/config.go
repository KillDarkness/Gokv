package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host         string
	Port         int
	Databases    int
	AppendOnly   bool
	AOFPath      string
	AOFFsync     string
	Snapshot     bool
	SnapshotPath string
	MaxKeys      int
	Eviction     string
}

func Default() Config {
	return Config{
		Host:         "0.0.0.0",
		Port:         6379,
		Databases:    1,
		AppendOnly:   false,
		AOFPath:      "data/appendonly.aof",
		AOFFsync:     "always",
		Snapshot:     false,
		SnapshotPath: "data/dump.gokv",
		MaxKeys:      0,
		Eviction:     "noeviction",
	}
}

func Load() (Config, error) {
	cfg := Default()

	if value := os.Getenv("GOKV_HOST"); value != "" {
		cfg.Host = value
	}
	if value := os.Getenv("GOKV_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid GOKV_PORT: %w", err)
		}
		cfg.Port = port
	}
	if value := os.Getenv("GOKV_DATABASES"); value != "" {
		databases, err := strconv.Atoi(value)
		if err != nil || databases < 1 {
			return Config{}, fmt.Errorf("invalid GOKV_DATABASES: %w", err)
		}
		cfg.Databases = databases
	}
	if value := os.Getenv("GOKV_APPENDONLY"); value != "" {
		appendOnly, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid GOKV_APPENDONLY: %w", err)
		}
		cfg.AppendOnly = appendOnly
	}
	if value := os.Getenv("GOKV_AOF_PATH"); value != "" {
		cfg.AOFPath = value
	}
	if value := os.Getenv("GOKV_AOF_FSYNC"); value != "" {
		cfg.AOFFsync = value
	}
	if value := os.Getenv("GOKV_SNAPSHOT"); value != "" {
		snapshot, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid GOKV_SNAPSHOT: %w", err)
		}
		cfg.Snapshot = snapshot
	}
	if value := os.Getenv("GOKV_SNAPSHOT_PATH"); value != "" {
		cfg.SnapshotPath = value
	}
	if value := os.Getenv("GOKV_MAXKEYS"); value != "" {
		maxKeys, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid GOKV_MAXKEYS: %w", err)
		}
		cfg.MaxKeys = maxKeys
	}
	if value := os.Getenv("GOKV_EVICTION"); value != "" {
		cfg.Eviction = value
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
