package mock

import "github.com/asdine/lobby"

var _ lobby.Registry = new(Registry)

// Registry is a mock service that runs provided functions. Useful for testing.
type Registry struct {
	CreateFn      func(string, string) error
	CreateInvoked int

	BucketFn      func(string) (lobby.Bucket, error)
	BucketInvoked int

	CloseFn      func() error
	CloseInvoked int
}

// RegisterBackend does nothing.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {}

// Create runs CreateFn and increments CreateInvoked when invoked.
func (r *Registry) Create(backendName, bucketName string) error {
	r.CreateInvoked++

	if r.CreateFn != nil {
		return r.CreateFn(backendName, bucketName)
	}

	return nil
}

// Bucket runs BucketFn and increments BucketInvoked when invoked.
func (r *Registry) Bucket(name string) (lobby.Bucket, error) {
	r.BucketInvoked++

	if r.BucketFn != nil {
		return r.BucketFn(name)
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
