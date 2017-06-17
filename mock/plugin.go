package mock

import "github.com/asdine/lobby"

var _ lobby.Plugin = new(Plugin)

// Plugin is a mock service that runs provided functions. Useful for testing.
type Plugin struct {
	NameFn      func() string
	NameInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Name runs NameFn and increments NameInvoked when invoked.
func (s *Plugin) Name() string {
	s.NameInvoked++

	if s.NameFn != nil {
		return s.NameFn()
	}

	return "mock"
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (s *Plugin) Close() error {
	s.CloseInvoked++

	if s.CloseFn != nil {
		return s.CloseFn()
	}

	return nil
}
