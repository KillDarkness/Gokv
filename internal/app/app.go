package app

import (
	"context"
	"time"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/config"
	appLog "github.com/KillDarkness/gokv/internal/log"
	"github.com/KillDarkness/gokv/internal/metrics"
	"github.com/KillDarkness/gokv/internal/persistence"
	"github.com/KillDarkness/gokv/internal/protocol"
	"github.com/KillDarkness/gokv/internal/server"
	"github.com/KillDarkness/gokv/internal/store"
)

type App struct {
	cfg      config.Config
	logger   *appLog.Logger
	registry *command.Registry
	server   *server.Server
	store    *store.Store
	aof      *persistence.AOF
	snapshot *persistence.Snapshot
	metrics  *metrics.Metrics
}

func New(cfg config.Config) *App {
	logger := appLog.New()
	st := store.New()
	metrics := metrics.New()
	aof, err := persistence.NewAOF(cfg.AppendOnly, cfg.AOFPath, cfg.AOFFsync)
	if err != nil {
		logger.Fatalf("invalid AOF config: %v", err)
	}
	snapshot := persistence.NewSnapshot(cfg.Snapshot, cfg.SnapshotPath)
	registry := command.NewDefaultRegistry(metrics)

	return &App{
		cfg:      cfg,
		logger:   logger,
		registry: registry,
		store:    st,
		aof:      aof,
		snapshot: snapshot,
		metrics:  metrics,
		server:   server.New(cfg, registry, st, aof, metrics, logger),
	}
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	janitorDone := a.store.StartJanitor(runCtx, time.Minute)
	aofSyncerDone := a.aof.StartSyncer(runCtx, time.Second)
	defer func() {
		cancel()
		<-janitorDone
		<-aofSyncerDone
		if err := a.snapshot.Save(context.Background(), a.store); err != nil {
			a.logger.Printf("snapshot save error: %v", err)
		}
		if err := a.aof.Rewrite(context.Background(), a.store); err != nil {
			a.logger.Printf("aof rewrite error: %v", err)
		}
		if err := a.aof.Close(); err != nil {
			a.logger.Printf("aof close error: %v", err)
		}
	}()
	if a.cfg.AppendOnly {
		a.logger.Printf("loading AOF from %s", a.cfg.AOFPath)
		if err := a.aof.Replay(runCtx, func(ctx context.Context, args []string) protocol.Reply {
			return a.registry.Dispatch(ctx, a.store, nil, args)
		}); err != nil {
			return err
		}
	} else if a.cfg.Snapshot {
		a.logger.Printf("loading snapshot from %s", a.cfg.SnapshotPath)
		if err := a.snapshot.Load(runCtx, a.store); err != nil {
			return err
		}
	}

	return a.server.ListenAndServe(runCtx)
}

func (a *App) Logger() *appLog.Logger {
	return a.logger
}
