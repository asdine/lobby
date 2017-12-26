package rpc_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryServerCreate(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, topicName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "topic", topicName)

			return nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewTopic{Name: "topic", Backend: "backend"})
		require.NoError(t, err)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), new(proto.NewTopic))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
	})

	t.Run("TopicAlreadyExists", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, topicName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "topic", topicName)

			return lobby.ErrTopicAlreadyExists
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewTopic{Name: "topic", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.AlreadyExists, grpc.Code(err))
	})

	t.Run("BackendNotFound", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, topicName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "topic", topicName)

			return lobby.ErrBackendNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewTopic{Name: "topic", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, topicName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "topic", topicName)

			return errors.New("something unexpected happened !")
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Create(context.Background(), &proto.NewTopic{Name: "topic", Backend: "backend"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func TestRegistryServerStatus(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return new(mock.Topic), nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		status, err := client.Status(context.Background(), &proto.Topic{Name: "topic"})
		require.NoError(t, err)
		require.True(t, status.Exists)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Status(context.Background(), new(proto.Topic))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
	})

	t.Run("NotFound", func(t *testing.T) {
		var r mock.Registry

		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return nil, lobby.ErrTopicNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		status, err := client.Status(context.Background(), &proto.Topic{Name: "topic"})
		require.NoError(t, err)
		require.False(t, status.Exists)
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry

		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return nil, errors.New("something unexpected happened !")
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewRegistryServiceClient(conn)

		_, err := client.Status(context.Background(), &proto.Topic{Name: "topic"})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}

func newRegistry(t *testing.T, r lobby.Registry) (*rpc.Registry, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	socketPath := path.Join(dir, "lobby.sock")
	l, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	srv := rpc.NewServer(log.New(log.Output(ioutil.Discard)), rpc.WithTopicService(r), rpc.WithRegistryService(r))

	go func() {
		srv.Serve(l)
	}()

	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}),
	)
	require.NoError(t, err)

	reg, err := rpc.NewRegistry(conn)
	require.NoError(t, err)
	reg.Backend = new(mock.Backend)

	return reg, func() {
		reg.Close()
		srv.Stop()
		os.RemoveAll(dir)
	}
}

func TestRegistryCreate(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.CreateFn = func(backendName, topicName string) error {
			assert.Equal(t, "backend", backendName)
			assert.Equal(t, "topic", topicName)

			return nil
		}

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		err := reg.Create("backend", "topic")
		require.NoError(t, err)
	})

	t.Run("Errors", func(t *testing.T) {
		var r mock.Registry

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		testCases := map[error]error{
			lobby.ErrTopicAlreadyExists: lobby.ErrTopicAlreadyExists,
			lobby.ErrBackendNotFound:    lobby.ErrBackendNotFound,
			lobby.ErrTopicNotFound:      lobby.ErrTopicNotFound,
			errors.New("unexpected"):    status.Error(codes.Unknown, rpc.ErrInternal.Error()),
		}

		for returnedErr, expectedErr := range testCases {
			testRegistryCreateWith(t, reg, &r, returnedErr, expectedErr)
		}
	})
}

func testRegistryCreateWith(t *testing.T, reg lobby.Registry, mockReg *mock.Registry, returnedErr, expectedErr error) {
	mockReg.CreateFn = func(backendName, topicName string) error {
		return returnedErr
	}

	err := reg.Create("backend", "topic")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestRegistryTopic(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return nil, nil
		}

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		_, err := reg.Topic("topic")
		require.NoError(t, err)
		require.Equal(t, 1, reg.Backend.(*mock.Backend).TopicInvoked)
	})

	t.Run("Errors", func(t *testing.T) {
		var r mock.Registry

		reg, cleanup := newRegistry(t, &r)
		defer cleanup()

		testCases := map[error]error{
			lobby.ErrTopicAlreadyExists: lobby.ErrTopicAlreadyExists,
			lobby.ErrBackendNotFound:    lobby.ErrBackendNotFound,
			lobby.ErrTopicNotFound:      lobby.ErrTopicNotFound,
			errors.New("unexpected"):    status.Error(codes.Unknown, rpc.ErrInternal.Error()),
		}

		for returnedErr, expectedErr := range testCases {
			testRegistryTopicWith(t, reg, &r, returnedErr, expectedErr)
		}
	})
}

func testRegistryTopicWith(t *testing.T, reg lobby.Registry, mockReg *mock.Registry, returnedErr, expectedErr error) {
	mockReg.TopicFn = func(name string) (lobby.Topic, error) {
		return nil, returnedErr
	}

	_, err := reg.Topic("topic")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}
