package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/KillDarkness/gokv/internal/app"
	"github.com/KillDarkness/gokv/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	application := app.New(cfg)
	if err := application.Run(ctx); err != nil {
		application.Logger().Fatalf("gokv stopped: %v", err)
	}
}
