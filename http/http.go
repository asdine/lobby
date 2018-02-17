package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/validation"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
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

// NewHandler instantiates a configured Handler.
func NewHandler(r lobby.Registry, logger *log.Logger) (http.Handler, error) {
	router := httprouter.New()

	h := handler{
		registry:     r,
		logger:       logger,
		router:       router,
		newEndpointC: make(chan lobby.Endpoint),
	}

	err := h.setupRouter()
	if err != nil {
		return nil, err
	}

	go func() {
		for range h.newEndpointC {
			err := h.setupRouter()
			if err != nil {
				logger.Printf("failed to refresh router, err: %s", err)
			}
		}
	}()

	return &wrapper{handler: router, logger: h.logger}, nil
}

type handler struct {
	sync.RWMutex

	registry     lobby.Registry
	router       *httprouter.Router
	logger       *log.Logger
	newEndpointC chan lobby.Endpoint
}

func (h *handler) setupRouter() error {
	h.Lock()
	defer h.Unlock()

	endpoints, err := h.registry.Endpoints()
	if err != nil {
		return errors.Wrap(err, "failed to fetch endpoints from registry")
	}

	router := httprouter.New()

	router.POST("/_/v1/endpoints", h.createEndpoint)

	for _, e := range endpoints {
		router.Handler(e.Method(), e.Path(), e)
	}

	h.router = router
	return nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RLock()
	h.router.ServeHTTP(w, r)
	h.RUnlock()
}

func (h *handler) createEndpoint(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req endpointCreationRequest

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

	endpoint, err := h.registry.Create(req.Backend, req.Path)
	if err != nil {
		switch err {
		case lobby.ErrBackendNotFound:
			http.NotFound(w, r)
		case lobby.ErrEndpointAlreadyExists:
			writeError(w, validation.AddError(nil, "path", err), http.StatusConflict, h.logger)
		default:
			writeError(w, err, http.StatusInternalServerError, h.logger)
		}
	}

	encodeJSON(w, &endpointCreationResponse{Path: endpoint.Path(), Backend: req.Backend}, http.StatusCreated, h.logger)
}

type endpointCreationRequest struct {
	Path    string `json:"path" valid:"required,stringlength(1|64)"`
	Backend string `json:"backend" valid:"required,alphanum"`
}

func (t *endpointCreationRequest) Validate() error {
	t.Backend = strings.TrimSpace(t.Backend)

	return validation.Validate(t)
}

type endpointCreationResponse struct {
	Path    string `json:"path"`
	Backend string `json:"backend"`
}
