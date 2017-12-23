package rpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestTopicServerSend(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var r mock.Registry

		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return &mock.Topic{
				SendFn: func(message *lobby.Message) error {
					assert.Equal(t, "group", message.Group)
					assert.Equal(t, "value", string(message.Value))
					return nil
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()

		client := proto.NewTopicServiceClient(conn)

		_, err := client.Send(context.Background(), &proto.NewMessage{
			Message: &proto.Message{
				Group: "group",
				Value: []byte("value"),
			},
			Topic: "topic",
		})
		require.NoError(t, err)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		var r mock.Registry
		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewTopicServiceClient(conn)

		_, err := client.Send(context.Background(), new(proto.NewMessage))
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, grpc.Code(err))
	})

	t.Run("TopicNotFound", func(t *testing.T) {
		var r mock.Registry
		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "unknown", name)
			return nil, lobby.ErrTopicNotFound
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewTopicServiceClient(conn)

		_, err := client.Send(context.Background(), &proto.NewMessage{
			Message: &proto.Message{
				Group: "group",
				Value: []byte("value"),
			},
			Topic: "unknown",
		})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, grpc.Code(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		var r mock.Registry
		r.TopicFn = func(name string) (lobby.Topic, error) {
			assert.Equal(t, "topic", name)

			return &mock.Topic{
				SendFn: func(message *lobby.Message) error {
					assert.Equal(t, "group", message.Group)
					assert.Equal(t, "value", string(message.Value))
					return errors.New("something unexpected happened !")
				},
			}, nil
		}

		conn, cleanup := newServer(t, &r)
		defer cleanup()
		client := proto.NewTopicServiceClient(conn)

		_, err := client.Send(context.Background(), &proto.NewMessage{
			Message: &proto.Message{
				Group: "group",
				Value: []byte("value"),
			},
			Topic: "topic",
		})
		require.Error(t, err)
		require.Equal(t, codes.Unknown, grpc.Code(err))
	})
}
