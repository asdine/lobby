package rpc

import (
	"io"
	"log"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/json"
	"github.com/asdine/lobby/rpc/proto"
	"golang.org/x/net/context"
)

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
			return Error(err, s.logger)
		}

		itemCount++
		b, err := s.registry.Bucket(newItem.Bucket)
		if err != nil {
			return Error(err, s.logger)
		}

		data := json.ToValidJSONFromBytes(newItem.Item.Value)
		_, err = b.Save(newItem.Item.Key, data)
		if err != nil {
			return Error(err, s.logger)
		}
	}
}

func (s *bucketService) Get(ctx context.Context, key *proto.Key) (*proto.Item, error) {
	return nil, nil
}

func (s *bucketService) Delete(ctx context.Context, key *proto.Key) (*proto.Empty, error) {
	return nil, nil
}

func (s *bucketService) List(page *proto.Page, stream proto.BucketService_ListServer) error {
	return nil
}
