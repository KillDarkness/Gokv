package server

import (
	"context"
	"errors"
	"io"

	"github.com/KillDarkness/gokv/internal/protocol"
)

func (s *Server) handle(ctx context.Context, reader io.Reader, writer io.Writer) {
	parser := protocol.NewParser(reader)

	for {
		args, err := parser.ReadCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			_ = protocol.WriteReply(writer, protocol.Error(err.Error()))
			return
		}

		reply := s.registry.Dispatch(ctx, s.store, s.appender, args)
		if err := protocol.WriteReply(writer, reply); err != nil {
			return
		}
	}
}
