package http

import (
	"context"
	"net/http"
	"time"
)

// Server wraps an HTTP server.
type Server struct {
	http.Server

	// Handler to serve.
	Handler http.Handler
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}
