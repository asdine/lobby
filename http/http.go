package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/asdine/lobby"
	ljson "github.com/asdine/lobby/json"
	"github.com/asdine/lobby/validation"
	"github.com/julienschmidt/httprouter"
)

const maxBodySize = 1024 * 1024

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

	if r.ContentLength > maxBodySize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)

	} else {
		s.ServeMux.ServeHTTP(rw, r)
	}

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

// encodeJSON encodes v to w in JSON format. Error() is called if encoding fails.
func encodeJSON(w http.ResponseWriter, v interface{}, status int, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		Error(w, err, http.StatusInternalServerError, logger)
	}
}

// NewHandler instantiates a configured Handler.
func NewHandler(r lobby.Registry) http.Handler {
	router := httprouter.New()

	handler := Handler{
		registry: r,
		logger:   log.New(os.Stderr, "", log.LstdFlags),
		router:   router,
	}

	router.POST("/v1/buckets/:backend", handler.createBucket)
	router.PUT("/v1/b/:bucket/:key", handler.putItem)
	return router
}

// Handler is the main http handler.
type Handler struct {
	registry lobby.Registry
	router   *httprouter.Router
	logger   *log.Logger
}

func (h *Handler) createBucket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req BucketCreationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		Error(w, ErrInvalidJSON, http.StatusBadRequest, h.logger)
		return
	}

	err = req.Validate()
	if err != nil {
		Error(w, err, http.StatusBadRequest, h.logger)
		return
	}

	err = h.registry.Create(ps.ByName("backend"), req.Name)
	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case lobby.ErrBackendNotFound:
		http.NotFound(w, r)
	case lobby.ErrBucketAlreadyExists:
		Error(w, validation.AddError(nil, "name", err), http.StatusBadRequest, h.logger)
	default:
		Error(w, err, http.StatusInternalServerError, h.logger)
	}
}

func (h *Handler) putItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.ContentLength == 0 {
		Error(w, ErrEmptyContent, http.StatusBadRequest, h.logger)
		return
	}

	data := ljson.ToValidJSONFromReader(r.Body)
	if len(data) == 0 {
		Error(w, ErrEmptyContent, http.StatusBadRequest, h.logger)
		return
	}
	defer r.Body.Close()

	b, err := h.registry.Bucket(ps.ByName("bucket"))
	if err != nil {
		if err == lobby.ErrBucketNotFound {
			http.NotFound(w, r)
			return
		}

		Error(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	item, err := b.Save(ps.ByName("key"), data)
	if err != nil {
		Error(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	encodeJSON(w, item, http.StatusOK, h.logger)
}

// BucketCreationRequest is used to create a bucket.
type BucketCreationRequest struct {
	Name string `json:"name" valid:"required,alphanum,stringlength(1|64)"`
}

// Validate bucket creation payload.
func (b *BucketCreationRequest) Validate() error {
	b.Name = strings.TrimSpace(b.Name)

	return validation.Validate(b)
}
