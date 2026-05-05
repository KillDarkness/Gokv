package app

import (
	"context"
	"time"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/config"
	appLog "github.com/KillDarkness/gokv/internal/log"
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
}

func New(cfg config.Config) *App {
	logger := appLog.New()
	st := store.New()
	aof := persistence.NewAOF(cfg.AppendOnly, cfg.AOFPath)
	registry := command.NewDefaultRegistry()

	return &App{
		cfg:      cfg,
		logger:   logger,
		registry: registry,
		store:    st,
		aof:      aof,
		server:   server.New(cfg, registry, st, aof, logger),
	}
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	janitorDone := a.store.StartJanitor(runCtx, time.Minute)
	defer func() {
		cancel()
		<-janitorDone
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
	}

	return a.server.ListenAndServe(runCtx)
}

func (a *App) Logger() *appLog.Logger {
	return a.logger
}
