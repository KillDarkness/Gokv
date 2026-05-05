package server

import (
	"bufio"
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
	bufferedWriter := bufio.NewWriterSize(writer, 32*1024)
	defer bufferedWriter.Flush()
	selectedDB := 0
	appender := newDatabaseAppender(s.appender)

	for {
		args, err := parser.ReadCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			_ = protocol.WriteReply(bufferedWriter, protocol.Error(err.Error()))
			_ = bufferedWriter.Flush()
			return
		}

		if strings.EqualFold(args[0], "SELECT") {
			reply := s.selectDatabase(args, &selectedDB)
			if err := protocol.WriteReply(bufferedWriter, reply); err != nil {
				return
			}
			if err := bufferedWriter.Flush(); err != nil {
				return
			}
			continue
		}

		appender.Select(selectedDB)
		reply := s.registry.Dispatch(ctx, s.stores[selectedDB], appender, args)
		if err := protocol.WriteReply(bufferedWriter, reply); err != nil {
			return
		}
		if err := bufferedWriter.Flush(); err != nil {
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
	lastDB   int
	hasDB    bool
	appender command.Appender
}

func newDatabaseAppender(appender command.Appender) *databaseAppender {
	return &databaseAppender{appender: appender}
}

func (a *databaseAppender) Select(db int) {
	a.db = db
}

func (a *databaseAppender) Append(ctx context.Context, args []string) error {
	if a.appender == nil {
		return nil
	}
	if !a.hasDB || a.lastDB != a.db {
		if err := a.appender.Append(ctx, []string{"SELECT", strconv.Itoa(a.db)}); err != nil {
			return err
		}
		a.lastDB = a.db
		a.hasDB = true
	}
	return a.appender.Append(ctx, args)
}

type staticDatabaseAppender struct {
	db       int
	appender command.Appender
}

func (a staticDatabaseAppender) Append(ctx context.Context, args []string) error {
	if a.appender == nil {
		return nil
	}
	if err := a.appender.Append(ctx, []string{"SELECT", strconv.Itoa(a.db)}); err != nil {
		return err
	}
	return a.appender.Append(ctx, args)
}
