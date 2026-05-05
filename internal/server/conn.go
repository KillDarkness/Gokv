package server

import (
	"context"
	"net"
)

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	s.metrics.IncConnections()
	defer s.metrics.DecConnections()
	s.handle(ctx, conn, conn)
}
