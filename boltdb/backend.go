package boltdb

import (
	"time"

	"github.com/asdine/brazier"
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/protobuf"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

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
		return nil, errors.Wrap(err, "Can't open database")
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
func (s *Backend) Bucket(name string) (brazier.Bucket, error) {
	return NewBucket(s.DB.From(name)), nil
}

// Close BoltDB connection.
func (s *Backend) Close() error {
	return s.DB.Close()
}
