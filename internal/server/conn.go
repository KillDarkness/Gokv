package server

import (
	"context"
	"net"
)

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.handle(ctx, conn, conn)
}
