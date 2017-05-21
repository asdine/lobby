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

func newBucketService(r lobby.Registry) *bucketService {
	return &bucketService{
		registry: r,
		logger:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

type bucketService struct {
	registry lobby.Registry
	logger   *log.Logger
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

		b, err := s.registry.Bucket(newItem.Bucket)
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

	b, err := s.registry.Bucket(key.Bucket)
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

	b, err := s.registry.Bucket(key.Bucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = b.Delete(key.Key)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}

func (s *bucketService) List(page *proto.Page, stream proto.BucketService_ListServer) error {
	err := validation.Validate(page)
	if err != nil {
		return newError(err, s.logger)
	}

	b, err := s.registry.Bucket(page.Bucket)
	if err != nil {
		return newError(err, s.logger)
	}

	p := 1
	pp := 20

	if page.Page > 0 {
		p = int(page.Page)
	}
	if page.PerPage > 0 {
		pp = int(page.PerPage)
	}

	items, err := b.Page(p, pp)
	if err != nil {
		return newError(err, s.logger)
	}

	for i := range items {
		err = stream.Send(&proto.Item{
			Key:   items[i].Key,
			Value: items[i].Value,
		})
		if err != nil {
			return newError(err, s.logger)
		}
	}
	return nil
}
