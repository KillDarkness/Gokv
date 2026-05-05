package app

import (
	"context"
	"time"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/config"
	appLog "github.com/KillDarkness/gokv/internal/log"
	"github.com/KillDarkness/gokv/internal/server"
	"github.com/KillDarkness/gokv/internal/store"
)

type App struct {
	cfg    config.Config
	logger *appLog.Logger
	server *server.Server
	store  *store.Store
}

func New(cfg config.Config) *App {
	logger := appLog.New()
	st := store.New()
	registry := command.NewDefaultRegistry()

	return &App{
		cfg:    cfg,
		logger: logger,
		store:  st,
		server: server.New(cfg, registry, st, logger),
	}
}

func (a *App) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	janitorDone := a.store.StartJanitor(runCtx, time.Minute)
	defer func() {
		cancel()
		<-janitorDone
	}()

	return a.server.ListenAndServe(runCtx)
}

func (a *App) Logger() *appLog.Logger {
	return a.logger
}
