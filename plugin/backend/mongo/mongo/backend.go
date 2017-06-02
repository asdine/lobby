package mongo

import (
	"github.com/asdine/lobby"
	"gopkg.in/mgo.v2"
)

var _ lobby.Backend = new(Backend)

// NewBackend returns a MongoDB backend.
func NewBackend(uri string) (*Backend, error) {
	var err error

	session, err := mgo.Dial(uri)
	if err != nil {
		return nil, err
	}

	err = ensureIndexes(session.DB(""))
	if err != nil {
		return nil, err
	}

	return &Backend{
		session: session,
	}, nil
}

func ensureIndexes(db *mgo.Database) error {
	col := db.C(colItems)

	index := mgo.Index{
		Key:    []string{"key"},
		Unique: true,
		Sparse: true,
	}

	return col.EnsureIndex(index)
}

// Backend is a MongoDB backend.
type Backend struct {
	session *mgo.Session
}

// Bucket returns the bucket associated with the given id.
func (s *Backend) Bucket(name string) (lobby.Bucket, error) {
	return NewBucket(s.session.Clone()), nil
}

// Close MongoDB connection.
func (s *Backend) Close() error {
	s.session.Close()
	return nil
}
