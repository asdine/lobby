package rpc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/require"
)

func TestPut(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		var i int
		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, value []byte) (*lobby.Item, error) {
					require.Equal(t, fmt.Sprintf("key%d", i+1), key)
					require.Equal(t, fmt.Sprintf(`"value%d"`, i+1), string(value))
					i++
					return &lobby.Item{
						Key:   key,
						Value: value,
					}, nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		stream, err := client.Put(context.Background())
		require.NoError(t, err)

		for j := 0; j < 5; j++ {
			err = stream.Send(&proto.NewItem{
				Item: &proto.Item{
					Key:   fmt.Sprintf("key%d", j+1),
					Value: []byte(fmt.Sprintf("value%d", j+1)),
				},
				Bucket: "bucket",
			})
			require.NoError(t, err)
		}

		summary, err := stream.CloseAndRecv()
		require.NoError(t, err)
		require.Equal(t, int32(5), summary.ItemCount)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewBucketServiceClient(conn)

		stream, err := client.Put(context.Background())
		require.NoError(t, err)

		err = stream.Send(new(proto.NewItem))
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		require.Error(t, err)
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		var r mock.Registry
		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "unknown", name)
			return nil, lobby.ErrBucketNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewBucketServiceClient(conn)

		stream, err := client.Put(context.Background())
		require.NoError(t, err)

		err = stream.Send(&proto.NewItem{
			Item: &proto.Item{
				Key:   "hello",
				Value: []byte("value"),
			},
			Bucket: "unknown",
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		require.Error(t, err)
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry
		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, data []byte) (*lobby.Item, error) {
					require.Equal(t, "hello", key)
					require.Equal(t, `"value"`, string(data))
					return nil, errors.New("something unexpected happened !")
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewBucketServiceClient(conn)

		stream, err := client.Put(context.Background())
		require.NoError(t, err)

		err = stream.Send(&proto.NewItem{
			Item: &proto.Item{
				Key:   "hello",
				Value: []byte("value"),
			},
			Bucket: "bucket",
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		require.Error(t, err)
	})
}
