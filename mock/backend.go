package mock

import "github.com/asdine/lobby"

var _ lobby.Backend = new(Backend)

// Backend is a mock service that runs provided functions. Useful for testing.
type Backend struct {
	EndpointFn      func(path string) (lobby.Endpoint, error)
	EndpointInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Endpoint runs EndpointFn and increments EndpointInvoked when invoked.
func (s *Backend) Endpoint(path string) (lobby.Endpoint, error) {
	s.EndpointInvoked++

	if s.EndpointFn != nil {
		return s.EndpointFn(path)
	}

	return nil, nil
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (s *Backend) Close() error {
	s.CloseInvoked++

	if s.CloseFn != nil {
		return s.CloseFn()
	}

	return nil
}
