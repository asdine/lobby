package rpc

import (
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/asdine/lobby/validation"
	"golang.org/x/net/context"
)

func newRegistryService(r lobby.Registry) *registryService {
	return &registryService{
		registry: r,
		logger:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

type registryService struct {
	registry lobby.Registry
	logger   *log.Logger
}

// Create a bucket in the registry.
func (s *registryService) Create(ctx context.Context, newBucket *proto.NewBucket) (*proto.Empty, error) {
	err := validation.Validate(newBucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = s.registry.Create(newBucket.Backend, newBucket.Name)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}

// Exists check a bucket in the registry.
func (s *registryService) Status(ctx context.Context, bucket *proto.Bucket) (*proto.BucketStatus, error) {
	err := validation.Validate(bucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	var exists bool

	_, err = s.registry.Bucket(bucket.Name)
	if err != nil {
		if err != lobby.ErrBucketNotFound {
			return nil, newError(err, s.logger)
		}
	} else {
		exists = true
	}

	return &proto.BucketStatus{
		Exists: exists,
	}, nil
}
