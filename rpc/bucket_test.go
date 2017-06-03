package rpc_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestBucketServerPut(t *testing.T) {
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
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
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
		require.Equal(t, codes.NotFound, grpc.Code(err))
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
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func TestBucketServerGet(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					require.Equal(t, "hello", key)

					return &lobby.Item{
						Key:   key,
						Value: []byte(`"value"`),
					}, nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		item, err := client.Get(context.Background(), &proto.Key{Bucket: "bucket", Key: "hello"})
		require.NoError(t, err)
		require.Equal(t, "hello", item.Key)
		require.Equal(t, `"value"`, string(item.Value))
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewBucketServiceClient(conn)

		_, err := client.Get(context.Background(), new(proto.Key))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
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

		_, err := client.Get(context.Background(), &proto.Key{Bucket: "unknown", Key: "hello"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					require.Equal(t, "unknown", key)

					return nil, lobby.ErrKeyNotFound
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		_, err := client.Get(context.Background(), &proto.Key{Bucket: "bucket", Key: "unknown"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					require.Equal(t, "unknown", key)

					return nil, errors.New("something unexpected happened !")
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		_, err := client.Get(context.Background(), &proto.Key{Bucket: "bucket", Key: "unknown"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func TestBucketServerDelete(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					require.Equal(t, "hello", key)

					return nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		_, err := client.Delete(context.Background(), &proto.Key{Bucket: "bucket", Key: "hello"})
		require.NoError(t, err)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewBucketServiceClient(conn)

		_, err := client.Delete(context.Background(), new(proto.Key))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
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

		_, err := client.Delete(context.Background(), &proto.Key{Bucket: "unknown", Key: "hello"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					require.Equal(t, "unknown", key)

					return lobby.ErrKeyNotFound
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		_, err := client.Delete(context.Background(), &proto.Key{Bucket: "bucket", Key: "unknown"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					require.Equal(t, "unknown", key)

					return errors.New("something unexpected happened !")
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		_, err := client.Delete(context.Background(), &proto.Key{Bucket: "bucket", Key: "unknown"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func TestBucketServerList(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PageFn: func(page, perPage int) ([]lobby.Item, error) {
					require.Equal(t, 10, page)
					require.Equal(t, 30, perPage)

					items := make([]lobby.Item, 5)
					for i := 0; i < 5; i++ {
						items[i].Key = fmt.Sprintf("key%d", i+1)
						items[i].Value = []byte(fmt.Sprintf(`"value%d"`, i+1))
					}
					return items, nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		stream, err := client.List(context.Background(), &proto.Page{Bucket: "bucket", Page: 10, PerPage: 30})
		require.NoError(t, err)

		var i int
		for {
			item, err := stream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("key%d", i+1), item.Key)
			require.Equal(t, fmt.Sprintf(`"value%d"`, i+1), string(item.Value))
			i++
		}
		require.Equal(t, 5, i)
	})

	t.Run("BadPages", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PageFn: func(page, perPage int) ([]lobby.Item, error) {
					require.Equal(t, 1, page)
					require.Equal(t, 20, perPage)

					return []lobby.Item{}, nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		stream, err := client.List(context.Background(), &proto.Page{Bucket: "bucket", Page: -10, PerPage: -30})
		require.NoError(t, err)

		_, err = stream.Recv()
		require.Equal(t, io.EOF, err)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		stream, err := client.List(context.Background(), new(proto.Page))
		require.NoError(t, err)

		_, err = stream.Recv()
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
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

		stream, err := client.List(context.Background(), &proto.Page{Bucket: "unknown"})
		require.NoError(t, err)

		_, err = stream.Recv()
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PageFn: func(page, perPage int) ([]lobby.Item, error) {
					require.Equal(t, 1, page)
					require.Equal(t, 20, perPage)

					return nil, errors.New("something unexpected happened !")
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewBucketServiceClient(conn)

		stream, err := client.List(context.Background(), &proto.Page{Bucket: "bucket"})
		require.NoError(t, err)

		_, err = stream.Recv()
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}
