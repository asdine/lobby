package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/asdine/lobby"
	ljson "github.com/asdine/lobby/json"
	"github.com/asdine/lobby/validation"
	"github.com/julienschmidt/httprouter"
)

const maxBodySize = 1024 * 1024

// NewServer returns an http lobby server.
func NewServer(handler http.Handler) lobby.Server {
	return &Server{
		Server: &http.Server{
			Handler: handler,
		},
	}
}

// Server wraps an HTTP server.
type Server struct {
	*http.Server
}

// Name of the server.
func (s *Server) Name() string {
	return "http"
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

type wrapper struct {
	handler http.Handler
	logger  *log.Logger
}

// ServeHTTP delegates a request to the underlying handler.
func (s *wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rw := newResponseWriter(w)

	if r.ContentLength > maxBodySize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	} else {
		s.handler.ServeHTTP(rw, r)
	}

	s.logger.Printf(
		"%s %s %s %d %d %s",
		clientIP(r),
		r.Method,
		r.URL,
		rw.status,
		rw.len,
		time.Since(start),
	)
}

// newResponseWriter instantiates a responseWriter.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// responseWriter is a wrapper around http.ResponseWriter.
// It allows to capture informations about the response.
type responseWriter struct {
	http.ResponseWriter

	status int
	len    int
}

// WriteHeader stores the status before calling the underlying
// http.ResponseWriter WriteHeader.
func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.len = len(data)
	return w.ResponseWriter.Write(data)
}

// encodeJSON encodes v to w in JSON format. Error() is called if encoding fails.
func encodeJSON(w http.ResponseWriter, v interface{}, status int, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		writeError(w, err, http.StatusInternalServerError, logger)
	}
}

// encodeJSON encodes v to w in JSON format. Error() is called if encoding fails.
func writeRawJSON(w http.ResponseWriter, v []byte, status int, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(v); err != nil {
		writeError(w, err, http.StatusInternalServerError, logger)
	}
}

// NewHandler instantiates a configured Handler.
func NewHandler(r lobby.Registry) http.Handler {
	router := httprouter.New()

	h := handler{
		registry: r,
		logger:   log.New(os.Stderr, "", log.LstdFlags),
		router:   router,
	}

	router.POST("/v1/buckets/:backend", h.createBucket)
	router.PUT("/v1/b/:bucket/:key", h.putItem)
	router.GET("/v1/b/:bucket/:key", h.getItem)
	router.DELETE("/v1/b/:bucket/:key", h.deleteItem)
	router.GET("/v1/b/:bucket", h.listItems)
	return &wrapper{handler: router, logger: h.logger}
}

// Handler is the main http handler.
type handler struct {
	registry lobby.Registry
	router   *httprouter.Router
	logger   *log.Logger
}

func (h *handler) createBucket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req bucketCreationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, errInvalidJSON, http.StatusBadRequest, h.logger)
		return
	}

	err = req.Validate()
	if err != nil {
		writeError(w, err, http.StatusBadRequest, h.logger)
		return
	}

	err = h.registry.Create(ps.ByName("backend"), req.Name)
	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case lobby.ErrBackendNotFound:
		http.NotFound(w, r)
	case lobby.ErrBucketAlreadyExists:
		writeError(w, validation.AddError(nil, "name", err), http.StatusBadRequest, h.logger)
	default:
		writeError(w, err, http.StatusInternalServerError, h.logger)
	}
}

func (h *handler) putItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.ContentLength == 0 {
		writeError(w, errEmptyContent, http.StatusBadRequest, h.logger)
		return
	}

	data := ljson.ToValidJSONFromReader(r.Body)
	if len(data) == 0 {
		writeError(w, errEmptyContent, http.StatusBadRequest, h.logger)
		return
	}
	defer r.Body.Close()

	b, err := h.registry.Bucket(ps.ByName("bucket"))
	if err != nil {
		if err == lobby.ErrBucketNotFound {
			http.NotFound(w, r)
			return
		}

		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	item, err := b.Put(ps.ByName("key"), data)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeRawJSON(w, item.Value, http.StatusOK, h.logger)
}

func (h *handler) getItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	b, err := h.registry.Bucket(ps.ByName("bucket"))
	if err != nil {
		if err == lobby.ErrBucketNotFound {
			http.NotFound(w, r)
			return
		}

		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	item, err := b.Get(ps.ByName("key"))
	switch err {
	case nil:
		writeRawJSON(w, item.Value, http.StatusOK, h.logger)
	case lobby.ErrKeyNotFound:
		http.NotFound(w, r)
	default:
		writeError(w, err, http.StatusInternalServerError, h.logger)
	}
}

func (h *handler) deleteItem(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	b, err := h.registry.Bucket(ps.ByName("bucket"))
	if err != nil {
		if err == lobby.ErrBucketNotFound {
			http.NotFound(w, r)
			return
		}

		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = b.Delete(ps.ByName("key"))
	switch err {
	case nil:
		w.WriteHeader(http.StatusNoContent)
	case lobby.ErrKeyNotFound:
		http.NotFound(w, r)
	default:
		writeError(w, err, http.StatusInternalServerError, h.logger)
	}
}

func (h *handler) listItems(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	b, err := h.registry.Bucket(ps.ByName("bucket"))
	if err != nil {
		if err == lobby.ErrBucketNotFound {
			http.NotFound(w, r)
			return
		}

		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage <= 0 {
		perPage = 20
	}

	items, err := b.Page(page, perPage)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	data, err := ljson.MarshalList(items)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeRawJSON(w, data, http.StatusOK, h.logger)
}

// BucketCreationRequest is used to create a bucket.
type bucketCreationRequest struct {
	Name string `json:"name" valid:"required,alphanum,stringlength(1|64)"`
}

// Validate bucket creation payload.
func (b *bucketCreationRequest) Validate() error {
	b.Name = strings.TrimSpace(b.Name)

	return validation.Validate(b)
}
