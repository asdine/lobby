package mock

import "github.com/asdine/lobby"

var _ lobby.Topic = new(Topic)

// Topic is a mock service that runs provided functions. Useful for testing.
type Topic struct {
	SendFn      func(*lobby.Message) error
	SendInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Send runs SendFn and increments SendInvoked when invoked.
func (b *Topic) Send(message *lobby.Message) error {
	b.SendInvoked++

	if b.SendFn != nil {
		return b.SendFn(message)
	}

	return nil
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (b *Topic) Close() error {
	b.CloseInvoked++

	if b.CloseFn != nil {
		return b.CloseFn()
	}

	return nil
}
