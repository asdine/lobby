package mock

import (
	"sync"

	"github.com/asdine/lobby"
)

var _ lobby.Plugin = new(Plugin)

// Plugin is a mock service that runs provided functions. Useful for testing.
type Plugin struct {
	sync.Mutex

	NameFn      func() string
	NameInvoked int

	CloseFn      func() error
	CloseInvoked int

	WaitFn      func() error
	WaitInvoked int
}

// Name runs NameFn and increments NameInvoked when invoked.
func (s *Plugin) Name() string {
	s.Lock()
	defer s.Unlock()
	s.NameInvoked++

	if s.NameFn != nil {
		return s.NameFn()
	}

	return "mock"
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (s *Plugin) Close() error {
	s.Lock()
	defer s.Unlock()
	s.CloseInvoked++

	if s.CloseFn != nil {
		return s.CloseFn()
	}

	return nil
}

// Wait runs WaitFn and increments WaitInvoked when invoked.
func (s *Plugin) Wait() error {
	s.Lock()
	defer s.Unlock()
	s.WaitInvoked++

	if s.WaitFn != nil {
		return s.WaitFn()
	}

	return nil
}
