package bolt

import (
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt/boltpb"
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/protobuf"
	"github.com/coreos/bbolt"
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

// Create a topic in the registry.
func (r *Registry) Create(backendName, topicName string) error {
	if _, ok := r.backends[backendName]; !ok {
		return lobby.ErrBackendNotFound
	}

	tx, err := r.DB.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create a transaction")
	}
	defer tx.Rollback()

	var topic boltpb.Topic

	err = tx.One("Name", topicName, &topic)
	if err == nil {
		return lobby.ErrTopicAlreadyExists
	}

	if err != storm.ErrNotFound {
		return errors.Wrapf(err, "failed to fetch topic %s", topicName)
	}

	err = tx.Save(&boltpb.Topic{
		Name:    topicName,
		Backend: backendName,
	})

	if err != nil {
		return errors.Wrapf(err, "failed to create topic %s", topicName)
	}

	err = tx.Commit()
	return errors.Wrap(err, "failed to commit")
}

// Topic returns the selected topic from the Backend.
func (r *Registry) Topic(name string) (lobby.Topic, error) {
	var topic boltpb.Topic

	err := r.DB.One("Name", name, &topic)
	if err == storm.ErrNotFound {
		return nil, lobby.ErrTopicNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch topic %s", name)
	}

	backend, ok := r.backends[topic.Backend]
	if !ok {
		return nil, lobby.ErrTopicNotFound
	}

	return backend.Topic(name)
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
