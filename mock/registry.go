package mock

import "github.com/asdine/lobby"

var _ lobby.Registry = new(Registry)

// Registry is a mock service that runs provided functions. Useful for testing.
type Registry struct {
	CreateFn      func(string, string) error
	CreateInvoked int

	TopicFn      func(string) (lobby.Topic, error)
	TopicInvoked int

	CloseFn      func() error
	CloseInvoked int

	Backends map[string]lobby.Backend
}

// RegisterBackend saves the backend in the Backends map.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {
	if r.Backends == nil {
		r.Backends = make(map[string]lobby.Backend)
	}

	r.Backends[name] = backend
}

// Create runs CreateFn and increments CreateInvoked when invoked.
func (r *Registry) Create(backendName, topicName string) error {
	r.CreateInvoked++

	if r.CreateFn != nil {
		return r.CreateFn(backendName, topicName)
	}

	return nil
}

// Topic runs TopicFn and increments TopicInvoked when invoked.
func (r *Registry) Topic(name string) (lobby.Topic, error) {
	r.TopicInvoked++

	if r.TopicFn != nil {
		return r.TopicFn(name)
	}

	return nil, nil
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (r *Registry) Close() error {
	r.CloseInvoked++

	if r.CloseFn != nil {
		return r.CloseFn()
	}

	return nil
}
