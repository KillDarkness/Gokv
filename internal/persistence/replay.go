package persistence

import "context"

func Replay(ctx context.Context) error {
	return ctx.Err()
}
