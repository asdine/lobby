package rpc_test

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBackend(t *testing.T, b lobby.Backend) (*rpc.Backend, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	socketPath := path.Join(dir, "lobby.sock")
	l, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	srv := rpc.NewServer(log.New(ioutil.Discard, ""), rpc.WithTopicService(b))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	backend, err := rpc.NewBackend(conn)
	require.NoError(t, err)

	return backend, func() {
		err := conn.Close()
		require.NoError(t, err)

		srv.Stop()
		wg.Wait()
		os.RemoveAll(dir)
	}
}

func TestTopicSend(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var b mock.Backend

		b.TopicFn = func(name string) (lobby.Topic, error) {
			require.Equal(t, "topic", name)

			return &mock.Topic{
				SendFn: func(message *lobby.Message) error {
					assert.Equal(t, "group", message.Group)
					assert.Equal(t, []byte(`Value`), message.Value)
					return nil
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		topic, err := backend.Topic("topic")
		require.NoError(t, err)

		err = topic.Send(&lobby.Message{
			Group: "group",
			Value: []byte("Value"),
		})
		require.NoError(t, err)
	})

	t.Run("TopicNotFound", func(t *testing.T) {
		var b mock.Backend
		b.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "unknown", name)
			return nil, lobby.ErrTopicNotFound
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		topic, err := backend.Topic("unknown")
		require.NoError(t, err)

		err = topic.Send(&lobby.Message{
			Group: "group",
			Value: []byte("Value"),
		})
		require.Error(t, err)
		require.Equal(t, lobby.ErrTopicNotFound, err)
	})

	t.Run("InternalError", func(t *testing.T) {
		var b mock.Backend

		b.TopicFn = func(name string) (lobby.Topic, error) {
			require.Equal(t, "topic", name)

			return &mock.Topic{
				SendFn: func(message *lobby.Message) error {
					assert.Equal(t, "group", message.Group)
					assert.Equal(t, []byte(`Value`), message.Value)
					return errors.New("something unexpected happened !")
				},
			}, nil
		}

		backend, cleanup := newBackend(t, &b)
		defer cleanup()

		topic, err := backend.Topic("topic")
		require.NoError(t, err)

		err = topic.Send(&lobby.Message{
			Group: "group",
			Value: []byte("Value"),
		})
		require.Error(t, err)
	})
}
