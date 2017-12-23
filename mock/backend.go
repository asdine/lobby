package mock

import "github.com/asdine/lobby"

var _ lobby.Backend = new(Backend)

// Backend is a mock service that runs provided functions. Useful for testing.
type Backend struct {
	TopicFn      func(name string) (lobby.Topic, error)
	TopicInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Topic runs TopicFn and increments TopicInvoked when invoked.
func (s *Backend) Topic(name string) (lobby.Topic, error) {
	s.TopicInvoked++

	if s.TopicFn != nil {
		return s.TopicFn(name)
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
