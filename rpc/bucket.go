package rpc

import (
	"io"
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/json"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/asdine/lobby/validation"
	"golang.org/x/net/context"
)

func newBucketService(b lobby.Backend) *bucketService {
	return &bucketService{
		backend: b,
		logger:  log.New(os.Stderr, "", log.LstdFlags),
	}
}

type bucketService struct {
	backend lobby.Backend
	logger  *log.Logger
}

// Put an item in the bucket.
func (s *bucketService) Put(stream proto.BucketService_PutServer) error {
	var itemCount int32

	for {
		newItem, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&proto.PutSummary{
				ItemCount: itemCount,
			})
		}
		if err != nil {
			return newError(err, s.logger)
		}

		itemCount++
		err = validation.Validate(newItem)
		if err != nil {
			return newError(err, s.logger)
		}

		b, err := s.backend.Bucket(newItem.Bucket)
		if err != nil {
			return newError(err, s.logger)
		}

		data := json.ToValidJSONFromBytes(newItem.Item.Value)
		_, err = b.Put(newItem.Item.Key, data)
		if err != nil {
			return newError(err, s.logger)
		}
	}
}

func (s *bucketService) Get(ctx context.Context, key *proto.Key) (*proto.Item, error) {
	err := validation.Validate(key)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	b, err := s.backend.Bucket(key.Bucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	item, err := b.Get(key.Key)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return &proto.Item{
		Key:   item.Key,
		Value: item.Value,
	}, nil
}

func (s *bucketService) Delete(ctx context.Context, key *proto.Key) (*proto.Empty, error) {
	err := validation.Validate(key)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	b, err := s.backend.Bucket(key.Bucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = b.Delete(key.Key)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}
