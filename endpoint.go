package lobby

import "io"

// Errors.
const (
	ErrBackendNotFound       = Error("backend not found")
	ErrEndpointNotFound      = Error("endpoint not found")
	ErrEndpointAlreadyExists = Error("endpoint already exists")
)

// Request ...
type Request struct {
	Header map[string]string
	Body   io.Reader
}

// An Endpoint ...
type Endpoint interface {
	// Send data to the endpoint.
	Send(*Request) error

	// Close the endpoint. Can be used to close sessions if required.
	Close() error
}

// EndpointFunc creates a endpoint from a send function.
func EndpointFunc(fn func(*Request) error) Endpoint {
	return &endpointFunc{fn}
}

type endpointFunc struct {
	fn func(*Request) error
}

func (t *endpointFunc) Send(r *Request) error {
	return t.fn(r)
}

func (t *endpointFunc) Close() error {
	return nil
}

// A Backend is able to create endpoints.
type Backend interface {
	// Get an endpoint by path.
	Endpoint(path string) (Endpoint, error)
	// Close the backend connection.
	Close() error
}

// A Registry manages the endpoints, their configuration and their associated Backend.
type Registry interface {
	Backend

	// Register a backend under the given name.
	RegisterBackend(name string, backend Backend)
	// Create a endpoint and register it to the Registry.
	Create(backendName, path string) error
}
