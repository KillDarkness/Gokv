package persistence

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/KillDarkness/gokv/internal/protocol"
)

func (a *AOF) Replay(ctx context.Context, apply func(context.Context, []string) protocol.Reply) error {
	if !a.Enabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := os.Open(a.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	parser := protocol.NewParser(file)
	for {
		args, err := parser.ReadCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		reply := apply(ctx, args)
		if replyErr, ok := reply.(protocol.Error); ok {
			return fmt.Errorf("replay command %q failed: %s", args[0], string(replyErr))
		}
	}
}
