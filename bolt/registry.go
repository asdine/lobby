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

// Registry is a BoltDB registry.
type Registry struct {
	DB       *storm.DB
	logger   *log.Logger
	backends map[string]lobby.Backend
}

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

// RegisterBackend registers a backend under the given name.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {
	r.backends[name] = backend
	r.logger.Debugf("Registered %s backend\n", name)
}

// Create a endpoint in the registry.
func (r *Registry) Create(backendName, path string) (lobby.Endpoint, error) {
	backend, ok := r.backends[backendName]
	if !ok {
		return nil, lobby.ErrBackendNotFound
	}

	tx, err := r.DB.Begin(true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a transaction")
	}
	defer tx.Rollback()

	var endpoint boltpb.Endpoint

	err = tx.One("Path", path, &endpoint)
	if err == nil {
		return nil, lobby.ErrEndpointAlreadyExists
	}

	if err != storm.ErrNotFound {
		return nil, errors.Wrapf(err, "failed to fetch endpoint %s", path)
	}

	err = tx.Save(&boltpb.Endpoint{
		Path:    path,
		Backend: backendName,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create endpoint %s", path)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction for endpoint %s", path)
	}

	return backend.Endpoint(path)
}

// Endpoint returns the selected endpoint from the Backend.
func (r *Registry) Endpoint(name string) (lobby.Endpoint, error) {
	var endpoint boltpb.Endpoint

	err := r.DB.One("Path", name, &endpoint)
	if err == storm.ErrNotFound {
		return nil, lobby.ErrEndpointNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch endpoint %s", name)
	}

	backend, ok := r.backends[endpoint.Backend]
	if !ok {
		return nil, lobby.ErrEndpointNotFound
	}

	return backend.Endpoint(name)
}

// Endpoints the endpoint list.
func (r *Registry) Endpoints() ([]lobby.Endpoint, error) {
	var endpoint boltpb.Endpoint

	var list []boltpb.Endpoint

	err := r.DB.All(&list)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all endpoints")
	}

	endpoints := make([]lobby.Endpoint, 0, len(list))
	for _, e := range list {
		backend, ok := r.backends[e.Backend]
		if !ok {
			return nil, errors.Errorf("backend not found for endpoint %s", endpoint.Path)
		}
		endpoint, err := backend.Endpoint(e.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get endpoint %s", endpoint.Path)
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
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
