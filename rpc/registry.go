package rpc

import (
	"context"
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/asdine/lobby/validation"
	"google.golang.org/grpc"
)

func newRegistryService(r lobby.Registry) *registryService {
	return &registryService{
		registry: r,
		logger:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

type registryService struct {
	registry lobby.Registry
	logger   *log.Logger
}

// Create a topic in the registry.
func (s *registryService) Create(ctx context.Context, newTopic *proto.NewTopic) (*proto.Empty, error) {
	err := validation.Validate(newTopic)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = s.registry.Create(newTopic.Backend, newTopic.Name)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}

// Exists check a topic in the registry.
func (s *registryService) Status(ctx context.Context, topic *proto.Topic) (*proto.TopicStatus, error) {
	err := validation.Validate(topic)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	var exists bool

	_, err = s.registry.Topic(topic.Name)
	if err != nil {
		if err != lobby.ErrTopicNotFound {
			return nil, newError(err, s.logger)
		}
	} else {
		exists = true
	}

	return &proto.TopicStatus{
		Exists: exists,
	}, nil
}

var _ lobby.Registry = new(Registry)

// NewRegistry returns a gRPC Registry. It is used to communicate with external Registries.
func NewRegistry(conn *grpc.ClientConn) (*Registry, error) {
	client := proto.NewRegistryServiceClient(conn)

	backend, err := NewBackend(conn)
	if err != nil {
		return nil, err
	}

	return &Registry{
		Backend: backend,
		conn:    conn,
		client:  client,
	}, nil
}

// Registry is a gRPC Registry.
type Registry struct {
	lobby.Backend

	conn   *grpc.ClientConn
	client proto.RegistryServiceClient
}

// RegisterBackend should never be called on this type.
func (s *Registry) RegisterBackend(_ string, _ lobby.Backend) {
	panic("RegisterBackend should not be called on this type")
}

// Create a topic and register it to the Registry.
func (s *Registry) Create(backendName, topicName string) error {
	_, err := s.client.Create(context.Background(), &proto.NewTopic{Name: topicName, Backend: backendName})
	return errFromGRPC(err)
}

// Topic returns the topic associated with the given id.
func (s *Registry) Topic(name string) (lobby.Topic, error) {
	status, err := s.client.Status(context.Background(), &proto.Topic{Name: name})
	if err != nil {
		return nil, errFromGRPC(err)
	}

	if !status.Exists {
		return nil, lobby.ErrTopicNotFound
	}

	return s.Backend.Topic(name)
}

// Close the connexion to the Registry.
func (s *Registry) Close() error {
	return s.conn.Close()
}
