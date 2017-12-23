package rpc

import (
	"context"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"google.golang.org/grpc"
)

var _ lobby.Backend = new(Backend)

// NewBackend returns a gRPC backend. It is used to communicate with external backends.
func NewBackend(conn *grpc.ClientConn) (*Backend, error) {
	client := proto.NewTopicServiceClient(conn)

	return &Backend{
		conn:   conn,
		client: client,
	}, nil
}

// Backend is a gRPC backend.
type Backend struct {
	conn   *grpc.ClientConn
	client proto.TopicServiceClient
}

// Topic returns the topic associated with the given name.
func (s *Backend) Topic(name string) (lobby.Topic, error) {
	return NewTopic(name, s.client), nil
}

// Close does nothing.
func (s *Backend) Close() error {
	return nil
}

var _ lobby.Topic = new(Topic)

// NewTopic returns a Topic.
func NewTopic(name string, client proto.TopicServiceClient) *Topic {
	return &Topic{
		name:   name,
		client: client,
	}
}

// Topic is a gRPC implementation of a topic.
type Topic struct {
	name   string
	client proto.TopicServiceClient
}

// Send a message to the topic.
func (t *Topic) Send(message *lobby.Message) error {
	_, err := t.client.Send(context.Background(), &proto.NewMessage{
		Topic: t.name,
		Message: &proto.Message{
			Group: message.Group,
			Value: message.Value,
		},
	})

	return errFromGRPC(err)
}

// Close the topic session.
func (t *Topic) Close() error {
	return nil
}
