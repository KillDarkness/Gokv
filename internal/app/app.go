package app

import (
	"context"
	"strconv"
	"strings"
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
	stores   []*store.Store
	aof      *persistence.AOF
	snapshot *persistence.Snapshot
	metrics  *metrics.Metrics
}

func New(cfg config.Config) *App {
	logger := appLog.New()
	stores := make([]*store.Store, cfg.Databases)
	for i := range stores {
		stores[i] = store.NewWithOptions(store.Options{MaxKeys: cfg.MaxKeys, EvictionPolicy: store.EvictionPolicy(cfg.Eviction)})
	}
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
		stores:   stores,
		aof:      aof,
		snapshot: snapshot,
		metrics:  metrics,
		server:   server.New(cfg, registry, stores, aof, metrics, logger),
	}
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	janitorDone := startJanitors(runCtx, a.stores, time.Minute)
	aofWriterDone := a.aof.StartWriter(runCtx, 4096)
	aofSyncerDone := a.aof.StartSyncer(runCtx, time.Second)
	defer func() {
		cancel()
		<-janitorDone
		<-aofWriterDone
		<-aofSyncerDone
		if err := a.snapshot.Save(context.Background(), a.stores[0]); err != nil {
			a.logger.Printf("snapshot save error: %v", err)
		}
		if err := a.aof.RewriteDatabases(context.Background(), a.stores); err != nil {
			a.logger.Printf("aof rewrite error: %v", err)
		}
		if err := a.aof.Close(); err != nil {
			a.logger.Printf("aof close error: %v", err)
		}
	}()
	if a.cfg.AppendOnly {
		a.logger.Printf("loading AOF from %s", a.cfg.AOFPath)
		selectedDB := 0
		if err := a.aof.Replay(runCtx, func(ctx context.Context, args []string) protocol.Reply {
			if len(args) == 2 && strings.EqualFold(args[0], "SELECT") {
				db, err := strconv.Atoi(args[1])
				if err != nil || db < 0 || db >= len(a.stores) {
					return protocol.Error("DB index is out of range")
				}
				selectedDB = db
				return protocol.SimpleString("OK")
			}
			return a.registry.Dispatch(ctx, a.stores[selectedDB], nil, args)
		}); err != nil {
			return err
		}
	} else if a.cfg.Snapshot {
		a.logger.Printf("loading snapshot from %s", a.cfg.SnapshotPath)
		if err := a.snapshot.Load(runCtx, a.stores[0]); err != nil {
			return err
		}
	}

	return a.server.ListenAndServe(runCtx)
}

func (a *App) Logger() *appLog.Logger {
	return a.logger
}

func startJanitors(ctx context.Context, stores []*store.Store, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		janitors := make([]<-chan struct{}, 0, len(stores))
		for _, st := range stores {
			janitors = append(janitors, st.StartJanitor(ctx, interval))
		}
		for _, janitorDone := range janitors {
			<-janitorDone
		}
	}()

	return done
}
