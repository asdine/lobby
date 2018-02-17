package lobby

import (
	"io"
	"net/http"
)

// Errors.
const (
	ErrBackendNotFound       = Error("backend not found")
	ErrEndpointNotFound      = Error("endpoint not found")
	ErrEndpointAlreadyExists = Error("endpoint already exists")
)

// An Endpoint ...
type Endpoint interface {
	io.Closer
	http.Handler

	Method() string
	Path() string
}

// EndpointFunc creates a endpoint from a send function.
// func EndpointFunc(fn func(*http.Request) error) Endpoint {
// 	return &endpointFunc{fn}
// }

// type endpointFunc struct {
// 	fn func(*http.Request) error
// }

// func (t *endpointFunc) Send(r *http.Request) error {
// 	return t.fn(r)
// }

// func (t *endpointFunc) Close() error {
// 	return nil
// }

// A Backend is able to create endpoints.
type Backend interface {
	io.Closer

	// Get an endpoint by path.
	Endpoint(path string) (Endpoint, error)
}

// A Registry manages the endpoints, their configuration and their associated Backend.
type Registry interface {
	Backend

	// Register a backend under the given name.
	RegisterBackend(name string, backend Backend)
	// Create a endpoint and register it to the Registry.
	Create(backendName, path string) (Endpoint, error)
	// List of all the endpoints paths.
	Endpoints() ([]Endpoint, error)
}
