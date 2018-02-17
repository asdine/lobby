package mock

import (
	"net/http"

	"github.com/asdine/lobby"
)

var _ lobby.Endpoint = new(Endpoint)

// Endpoint is a mock service that runs provided functions. Useful for testing.
type Endpoint struct {
	ServeHTTPFn      func(w http.ResponseWriter, r *http.Request)
	ServeHTTPInvoked int

	MethodInvoked int
	MethodValue   string

	PathInvoked int
	PathValue   string

	CloseFn      func() error
	CloseInvoked int
}

// ServeHTTP runs ServeHTTPFn and increments ServeHTTPInvoked when invoked.
func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.ServeHTTPInvoked++

	if e.ServeHTTPFn != nil {
		e.ServeHTTPFn(w, r)
	}
}

// Method returns the MethodValue and increments MethodInvoked when invoked.
func (e *Endpoint) Method() string {
	return e.MethodValue
}

// Path returns the PathValue and increments PathInvoked when invoked.
func (e *Endpoint) Path() string {
	return e.PathValue
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (e *Endpoint) Close() error {
	e.CloseInvoked++

	if e.CloseFn != nil {
		return e.CloseFn()
	}

	return nil
}
