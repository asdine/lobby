package mock

import "github.com/asdine/lobby"

var _ lobby.Bucket = new(Bucket)

// Bucket is a mock service that runs provided functions. Useful for testing.
type Bucket struct {
	PutFn      func(key string, value []byte) (*lobby.Item, error)
	PutInvoked int

	GetFn      func(key string) (*lobby.Item, error)
	GetInvoked int

	DeleteFn      func(key string) error
	DeleteInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Put runs PutFn and increments PutInvoked when invoked.
func (b *Bucket) Put(key string, value []byte) (*lobby.Item, error) {
	b.PutInvoked++

	if b.PutFn != nil {
		return b.PutFn(key, value)
	}

	return nil, nil
}

// Get runs GetFn and increments GetInvoked when invoked.
func (b *Bucket) Get(key string) (*lobby.Item, error) {
	b.GetInvoked++

	if b.GetFn != nil {
		return b.GetFn(key)
	}

	return nil, nil
}

// Delete runs DeleteFn and increments DeleteInvoked when invoked.
func (b *Bucket) Delete(key string) error {
	b.DeleteInvoked++

	if b.DeleteFn != nil {
		return b.DeleteFn(key)
	}

	return nil
}

// Close runs CloseFn and increments CloseInvoked when invoked.
func (b *Bucket) Close() error {
	b.CloseInvoked++

	if b.CloseFn != nil {
		return b.CloseFn()
	}

	return nil
}
