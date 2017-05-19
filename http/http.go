package http

import (
	"context"
	"log"
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

// NewServeMux instantiates a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{
		ServeMux: http.NewServeMux(),
	}
}

// ServeMux is a wrapper around a http.Handler.
type ServeMux struct {
	*http.ServeMux
}

// ServeHTTP delegates a request to the underlying ServeMux.
func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rw := NewResponseWriter(w)
	s.ServeMux.ServeHTTP(rw, r)

	log.Printf(
		"%s %s %s %d %d %s",
		clientIP(r),
		r.Method,
		r.URL,
		rw.status,
		rw.len,
		time.Since(start),
	)
}

// NewResponseWriter instantiates a ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// ResponseWriter is a wrapper around http.ResponseWriter.
// It allows to capture informations about the response.
type ResponseWriter struct {
	http.ResponseWriter

	status int
	len    int
}

// WriteHeader stores the status before calling the underlying
// http.ResponseWriter WriteHeader.
func (w *ResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.len = len(data)
	return w.ResponseWriter.Write(data)
}
