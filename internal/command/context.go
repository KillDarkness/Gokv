package command

import (
	"context"
	"time"

	"github.com/KillDarkness/gokv/internal/store"
)

type Context struct {
	Context   context.Context
	Store     *store.Store
	Args      []string
	StartedAt time.Time
}
