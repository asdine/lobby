package mock

import "github.com/asdine/lobby"

var _ lobby.Backend = new(Backend)

// Backend is a mock service that runs provided functions. Useful for testing.
type Backend struct {
	BucketFn      func(name string) (lobby.Bucket, error)
	BucketInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// Bucket runs BucketFn and increments BucketInvoked when invoked.
func (s *Backend) Bucket(name string) (lobby.Bucket, error) {
	s.BucketInvoked++

	if s.BucketFn != nil {
		return s.BucketFn(name)
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
