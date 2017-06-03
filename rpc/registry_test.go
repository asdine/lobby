package rpc_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryServerCreate(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewBucket{Name: "bucket", Backend: "backend"})
		require.NoError(t, err)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), new(proto.NewBucket))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
	})

	t.Run("BucketAlreadyExists", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return lobby.ErrBucketAlreadyExists
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewBucket{Name: "bucket", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.AlreadyExists, grpc.Code(err))
	})

	t.Run("BackendNotFound", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return lobby.ErrBackendNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewBucket{Name: "bucket", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return errors.New("something unexpected happened !")
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewBucket{Name: "bucket", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func TestRegistryServerStatus(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return new(mock.Bucket), nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		status, err := client.Status(context.Background(), &proto.Bucket{Name: "bucket"})
		require.NoError(t, err)
		require.True(t, status.Exists)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Status(context.Background(), new(proto.Bucket))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
	})

	t.Run("NotFound", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, lobby.ErrBucketNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		status, err := client.Status(context.Background(), &proto.Bucket{Name: "bucket"})
		require.NoError(t, err)
		require.False(t, status.Exists)
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, errors.New("something unexpected happened !")
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Status(context.Background(), &proto.Bucket{Name: "bucket"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func newRegistry(t *testing.T, r lobby.Registry) (*rpc.Registry, func()) {
	l, err := net.Listen("tcp", ":")
	require.NoError(t, err)

	srv := rpc.NewServer(rpc.WithBucketService(r), rpc.WithRegistryService(r))

	go func() {
		srv.Serve(l)
	}()

	conn, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	reg, err := rpc.NewRegistry(conn)
	require.NoError(t, err)
	reg.Backend = new(mock.Backend)

	return reg, func() {
		reg.Close()
		srv.Stop()
	}
}

func TestRegistryCreate(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, bucketName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "bucket", bucketName)

			return nil
		}

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		err := reg.Create("backend", "bucket")
		require.NoError(t, err)
	})

	t.Run("Errors", func(t *testing.T) {
		var r mock.Registry

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		testCases := map[error]error{
			lobby.ErrBucketAlreadyExists: lobby.ErrBucketAlreadyExists,
			lobby.ErrKeyNotFound:         lobby.ErrKeyNotFound,
			lobby.ErrBackendNotFound:     lobby.ErrBackendNotFound,
			lobby.ErrBucketNotFound:      lobby.ErrBucketNotFound,
			errors.New("unexpected"):     status.Error(codes.Unknown, rpc.ErrInternal.Error()),
		}

		for returnedErr, expectedErr := range testCases {
			testRegistryCreateWith(t, reg, &r, returnedErr, expectedErr)
		}
	})
}

func testRegistryCreateWith(t *testing.T, reg lobby.Registry, mockReg *mock.Registry, returnedErr, expectedErr error) {
	mockReg.CreateFn = func(backendName, bucketName string) error {
		return returnedErr
	}

	err := reg.Create("backend", "bucket")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestRegistryBucket(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.BucketFn = func(name string) (lobby.Bucket, error) {
			assert.Equal(t, "bucket", name)

			return nil, nil
		}

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		_, err := reg.Bucket("bucket")
		require.NoError(t, err)
		require.Equal(t, 1, reg.Backend.(*mock.Backend).BucketInvoked)
	})

	t.Run("Errors", func(t *testing.T) {
		var r mock.Registry

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		testCases := map[error]error{
			lobby.ErrBucketAlreadyExists: lobby.ErrBucketAlreadyExists,
			lobby.ErrKeyNotFound:         lobby.ErrKeyNotFound,
			lobby.ErrBackendNotFound:     lobby.ErrBackendNotFound,
			lobby.ErrBucketNotFound:      lobby.ErrBucketNotFound,
			errors.New("unexpected"):     status.Error(codes.Unknown, rpc.ErrInternal.Error()),
		}

		for returnedErr, expectedErr := range testCases {
			testRegistryBucketWith(t, reg, &r, returnedErr, expectedErr)
		}
	})
}

func testRegistryBucketWith(t *testing.T, reg lobby.Registry, mockReg *mock.Registry, returnedErr, expectedErr error) {
	mockReg.BucketFn = func(name string) (lobby.Bucket, error) {
		return nil, returnedErr
	}

	_, err := reg.Bucket("bucket")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}
