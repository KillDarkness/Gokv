package server

import (
	"context"
	"errors"
	"net"

	"github.com/KillDarkness/gokv/internal/command"
	"github.com/KillDarkness/gokv/internal/config"
	appLog "github.com/KillDarkness/gokv/internal/log"
	"github.com/KillDarkness/gokv/internal/store"
)

type Server struct {
	cfg      config.Config
	registry *command.Registry
	store    *store.Store
	appender command.Appender
	logger   *appLog.Logger
}

func New(cfg config.Config, registry *command.Registry, st *store.Store, appender command.Appender, logger *appLog.Logger) *Server {
	return &Server{cfg: cfg, registry: registry, store: st, appender: appender, logger: logger}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.cfg.Addr())
	if err != nil {
		return err
	}
	defer listener.Close()

	s.logger.Printf("listening on %s", s.cfg.Addr())

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil
			}
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.logger.Printf("accept error: %v", err)
			continue
		}

		go s.handleConn(ctx, conn)
	}
}
