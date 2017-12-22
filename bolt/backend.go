package bolt

import (
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/protobuf"
	"github.com/coreos/bbolt"
)

var _ lobby.Backend = new(Backend)

// NewBackend returns a BoltDB backend.
func NewBackend(path string) (*Backend, error) {
	var err error

	db, err := storm.Open(
		path,
		storm.Codec(protobuf.Codec),
		storm.BoltOptions(0644, &bolt.Options{
			Timeout: time.Duration(50) * time.Millisecond,
		}),
	)

	if err != nil {
		return nil, err
	}

	return &Backend{
		DB: db,
	}, nil
}

// Backend is a BoltDB backend.
type Backend struct {
	DB *storm.DB
}

// Bucket returns the bucket associated with the given id.
func (s *Backend) Bucket(name string) (lobby.Bucket, error) {
	return NewBucket(s.DB.From(name)), nil
}

// Close BoltDB connection.
func (s *Backend) Close() error {
	return s.DB.Close()
}
