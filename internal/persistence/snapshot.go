package persistence

import "context"

type Snapshot struct {
	Enabled bool
}

func (s *Snapshot) Save(ctx context.Context) error {
	return ctx.Err()
}
