package command

import (
	"context"
	"time"

	"github.com/KillDarkness/gokv/internal/metrics"
	"github.com/KillDarkness/gokv/internal/store"
)

type Appender interface {
	Append(ctx context.Context, args []string) error
}

type Context struct {
	Context   context.Context
	Store     *store.Store
	Appender  Appender
	Metrics   *metrics.Metrics
	Args      []string
	StartedAt time.Time
}
