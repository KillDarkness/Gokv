package persistence

import "context"

type AOF struct {
	Enabled bool
}

func (a *AOF) Append(ctx context.Context, args []string) error {
	return ctx.Err()
}
