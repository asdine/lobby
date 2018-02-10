package bolt

import (
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt/boltpb"
	"github.com/asdine/lobby/log"
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/protobuf"
	"github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

var _ lobby.Registry = new(Registry)

// NewRegistry returns a BoltDB Registry.
func NewRegistry(path string, logger *log.Logger) (*Registry, error) {
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
		logger:   logger,
		backends: make(map[string]lobby.Backend),
	}, nil
}

// Registry is a BoltDB registry.
type Registry struct {
	DB       *storm.DB
	logger   *log.Logger
	backends map[string]lobby.Backend
}

// RegisterBackend registers a backend under the given name.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {
	r.backends[name] = backend
	r.logger.Debugf("Registered %s backend\n", name)
}

// Create a topic in the registry.
func (r *Registry) Create(backendName, path string) error {
	if _, ok := r.backends[backendName]; !ok {
		return lobby.ErrBackendNotFound
	}

	tx, err := r.DB.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create a transaction")
	}
	defer tx.Rollback()

	var topic boltpb.Endpoint

	err = tx.One("Path", path, &topic)
	if err == nil {
		return lobby.ErrEndpointAlreadyExists
	}

	if err != storm.ErrNotFound {
		return errors.Wrapf(err, "failed to fetch topic %s", path)
	}

	err = tx.Save(&boltpb.Endpoint{
		Path:    path,
		Backend: backendName,
	})

	if err != nil {
		return errors.Wrapf(err, "failed to create topic %s", path)
	}

	err = tx.Commit()
	return errors.Wrap(err, "failed to commit")
}

// Endpoint returns the selected topic from the Backend.
func (r *Registry) Endpoint(name string) (lobby.Endpoint, error) {
	var topic boltpb.Endpoint

	err := r.DB.One("Path", name, &topic)
	if err == storm.ErrNotFound {
		return nil, lobby.ErrEndpointNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch topic %s", name)
	}

	backend, ok := r.backends[topic.Backend]
	if !ok {
		return nil, lobby.ErrEndpointNotFound
	}

	return backend.Endpoint(name)
}

// Close BoltDB connection and registered backends.
func (r *Registry) Close() error {
	for name, backend := range r.backends {
		err := backend.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to close backend %s", name)
		}

		r.logger.Debugf("Stopped %s backend\n", name)
	}

	err := r.DB.Close()

	return errors.Wrap(err, "failed to close registry")
}
