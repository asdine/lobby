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
	col := db.C(colMessages)

	index := mgo.Index{
		Key:    []string{"topic", "group"},
		Sparse: true,
	}

	return col.EnsureIndex(index)
}

// Backend is a MongoDB backend.
type Backend struct {
	session *mgo.Session
}

// Topic returns the topic associated with the given name.
func (s *Backend) Topic(name string) (lobby.Topic, error) {
	return NewTopic(s.session.Copy(), name), nil
}

// Close MongoDB connection.
func (s *Backend) Close() error {
	s.session.Close()
	return nil
}
