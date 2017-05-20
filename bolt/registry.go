package bolt

import (
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt/internal"
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/protobuf"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

var _ lobby.Registry = new(Registry)

// NewRegistry returns a BoltDB Registry.
func NewRegistry(path string) (*Registry, error) {
	var err error

	db, err := storm.Open(
		path,
		storm.Codec(protobuf.Codec),
		storm.BoltOptions(0644, &bolt.Options{
			Timeout: time.Duration(50) * time.Millisecond,
		}),
	)

	if err != nil {
		return nil, errors.Wrap(err, "Can't open database")
	}

	return &Registry{
		DB:       db,
		backends: make(map[string]lobby.Backend),
	}, nil
}

// Registry is a BoltDB registry.
type Registry struct {
	DB       *storm.DB
	backends map[string]lobby.Backend
}

// RegisterBackend registers a backend under the given name.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {
	r.backends[name] = backend
}

// Create a bucket in the registry.
func (r *Registry) Create(backendName, bucketName string) error {
	if _, ok := r.backends[backendName]; !ok {
		return lobby.ErrBackendNotFound
	}

	tx, err := r.DB.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create a transaction")
	}
	defer tx.Rollback()

	var bucket internal.Bucket

	err = tx.One("Name", bucketName, &bucket)
	if err == nil {
		return lobby.ErrBucketAlreadyExists
	}

	if err != storm.ErrNotFound {
		return errors.Wrapf(err, "failed to fetch bucket %s", bucketName)
	}

	err = tx.Save(&internal.Bucket{
		Name:    bucketName,
		Backend: backendName,
	})

	if err != nil {
		return errors.Wrapf(err, "failed to create bucket %s", bucketName)
	}

	err = tx.Commit()
	return errors.Wrap(err, "failed to commit")
}

// Bucket returns the selected bucket from the Backend.
func (r *Registry) Bucket(name string) (lobby.Bucket, error) {
	var bucket internal.Bucket

	err := r.DB.One("Name", name, &bucket)
	if err == storm.ErrNotFound {
		return nil, lobby.ErrBucketNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch bucket %s", name)
	}

	backend := r.backends[bucket.Backend]
	return backend.Bucket(name)
}

// Close BoltDB connection and registered backends.
func (r *Registry) Close() error {
	for name, backend := range r.backends {
		err := backend.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to close backend %s", name)
		}
	}

	err := r.DB.Close()

	return errors.Wrap(err, "failed to close registry")
}
