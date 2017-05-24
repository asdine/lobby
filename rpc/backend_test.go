package rpc_test

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBackend(t *testing.T, b lobby.Backend) (*rpc.Backend, func()) {
	l, err := net.Listen("tcp", ":")
	require.NoError(t, err)

	srv := rpc.NewServer(rpc.WithBucketService(b))

	go func() {
		srv.Serve(l)
	}()

	backend, err := rpc.NewBackend(l.Addr().String())
	require.NoError(t, err)

	return backend, func() {
		backend.Close()
		srv.Stop()
	}
}

func TestBucketPut(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, value []byte) (*lobby.Item, error) {
					assert.Equal(t, "key", key)
					assert.Equal(t, []byte(`"Value"`), value)
					return &lobby.Item{
						Key:   key,
						Value: value,
					}, nil
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		item, err := bucket.Put("key", []byte(`"Value"`))
		require.NoError(t, err)
		require.Equal(t, "key", item.Key)
		require.Equal(t, []byte(`"Value"`), item.Value)
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		var b mock.Backend
		b.BucketFn = func(name string) (lobby.Bucket, error) {
			assert.Equal(t, "unknown", name)
			return nil, lobby.ErrBucketNotFound
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("unknown")
		require.NoError(t, err)

		_, err = bucket.Put("key", []byte(`"Value"`))
		require.Error(t, err)
		require.Equal(t, lobby.ErrBucketNotFound, err)
	})

	t.Run("InternalError", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, value []byte) (*lobby.Item, error) {
					assert.Equal(t, "key", key)
					assert.Equal(t, []byte(`"Value"`), value)
					return nil, errors.New("something unexpected happened !")
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		_, err = bucket.Put("key", []byte(`"Value"`))
		require.Error(t, err)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, value []byte) (*lobby.Item, error) {
					assert.Equal(t, "key", key)
					assert.Equal(t, []byte(`"Value"`), value)
					return nil, lobby.ErrKeyNotFound
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		_, err = bucket.Put("key", []byte(`"Value"`))
		require.Error(t, err)
		require.Equal(t, lobby.ErrKeyNotFound, err)
	})
}

func TestBucketGet(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			assert.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					assert.Equal(t, "key", key)
					return &lobby.Item{
						Key:   key,
						Value: []byte(`"Value"`),
					}, nil
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		item, err := bucket.Get("key")
		require.NoError(t, err)
		require.Equal(t, "key", item.Key)
		require.Equal(t, []byte(`"Value"`), item.Value)
	})
}

func TestBucketDelete(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			assert.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					assert.Equal(t, "key", key)
					return nil
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		err = bucket.Delete("key")
		require.NoError(t, err)
	})
}

func TestBucketList(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var b mock.Backend

		b.BucketFn = func(name string) (lobby.Bucket, error) {
			assert.Equal(t, "bucket", name)

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

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		bucket, err := backend.Bucket("bucket")
		require.NoError(t, err)

		items, err := bucket.Page(10, 30)
		require.NoError(t, err)
		require.Len(t, items, 5)
	})
}
