package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
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

// Serve incoming requests.
func (s *Server) Serve(l net.Listener) error {
	err := s.Server.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
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

	s.logger.Debugf(
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
func NewHandler(r lobby.Registry, logger *log.Logger) http.Handler {
	router := httprouter.New()

	h := handler{
		registry: r,
		logger:   logger,
		router:   router,
	}

	router.POST("/v1/topics", h.createTopic)
	router.POST("/v1/topics/:topic", h.postMessage)
	router.POST("/v1/topics/:topic/:group", h.postMessage)
	return &wrapper{handler: router, logger: h.logger}
}

type handler struct {
	registry lobby.Registry
	router   *httprouter.Router
	logger   *log.Logger
}

func (h *handler) createTopic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req topicCreationRequest

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, nil, http.StatusUnsupportedMediaType, h.logger)
		return
	}

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

	err = h.registry.Create(req.Backend, req.Name)
	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case lobby.ErrBackendNotFound:
		http.NotFound(w, r)
	case lobby.ErrTopicAlreadyExists:
		writeError(w, validation.AddError(nil, "name", err), http.StatusBadRequest, h.logger)
	default:
		writeError(w, err, http.StatusInternalServerError, h.logger)
	}
}

func (h *handler) postMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.ContentLength == 0 {
		writeError(w, errEmptyContent, http.StatusBadRequest, h.logger)
		return
	}

	defer r.Body.Close()
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	t, err := h.registry.Topic(ps.ByName("topic"))
	if err != nil {
		if err == lobby.ErrTopicNotFound {
			http.NotFound(w, r)
			return
		}

		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	err = t.Send(&lobby.Message{
		Group: ps.ByName("group"),
		Value: value,
	})
	if err != nil {
		writeError(w, err, http.StatusInternalServerError, h.logger)
		return
	}

	writeRawJSON(w, nil, http.StatusCreated, h.logger)
}

type topicCreationRequest struct {
	Name    string `json:"name" valid:"required,alphanum,stringlength(1|64)"`
	Backend string `json:"backend" valid:"required,alphanum"`
}

func (t *topicCreationRequest) Validate() error {
	t.Name = strings.TrimSpace(t.Name)
	t.Backend = strings.TrimSpace(t.Backend)

	return validation.Validate(t)
}
