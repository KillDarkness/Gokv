package server

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/protocol"
)

func (s *Server) handle(ctx context.Context, reader io.Reader, writer io.Writer) {
	parser := protocol.NewParser(reader)
	selectedDB := 0

	for {
		args, err := parser.ReadCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			_ = protocol.WriteReply(writer, protocol.Error(err.Error()))
			return
		}

		if strings.EqualFold(args[0], "SELECT") {
			reply := s.selectDatabase(args, &selectedDB)
			if err := protocol.WriteReply(writer, reply); err != nil {
				return
			}
			continue
		}

		reply := s.registry.Dispatch(ctx, s.stores[selectedDB], databaseAppender{db: selectedDB, appender: s.appender}, args)
		if err := protocol.WriteReply(writer, reply); err != nil {
			return
		}
	}
}

func (s *Server) selectDatabase(args []string, selectedDB *int) protocol.Reply {
	if len(args) != 2 {
		return protocol.Error("wrong number of arguments for 'select' command")
	}
	db, err := strconv.Atoi(args[1])
	if err != nil || db < 0 || db >= len(s.stores) {
		return protocol.Error("DB index is out of range")
	}
	*selectedDB = db
	return protocol.SimpleString("OK")
}

type databaseAppender struct {
	db       int
	appender command.Appender
}

func (a databaseAppender) Append(ctx context.Context, args []string) error {
	if a.appender == nil {
		return nil
	}
	if err := a.appender.Append(ctx, []string{"SELECT", strconv.Itoa(a.db)}); err != nil {
		return err
	}
	return a.appender.Append(ctx, args)
}
