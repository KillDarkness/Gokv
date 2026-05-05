package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/KillDarkness/gokv/internal/app"
	"github.com/KillDarkness/gokv/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := app.New(config.Default())
	if err := application.Run(ctx); err != nil {
		application.Logger().Fatalf("gokv stopped: %v", err)
	}
}
