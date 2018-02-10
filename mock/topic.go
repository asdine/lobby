package mock

import "github.com/asdine/lobby"

var _ lobby.Endpoint = new(Endpoint)

// Endpoint is a mock service that runs provided functions. Useful for testing.
type Endpoint struct {
	SendFn      func(*lobby.Request) error
	SendInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Send runs SendFn and increments SendInvoked when invoked.
func (e *Endpoint) Send(message *lobby.Request) error {
	e.SendInvoked++

	if e.SendFn != nil {
		return e.SendFn(message)
	}

	return nil
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (e *Endpoint) Close() error {
	e.CloseInvoked++

	if e.CloseFn != nil {
		return e.CloseFn()
	}

	return nil
}
