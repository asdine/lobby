package rpc_test

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc/proto"
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
		require.Equal(t, codes.Internal, grpc.Code(err))
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
		require.Equal(t, codes.Internal, grpc.Code(err))
	})
}
