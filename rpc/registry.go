package rpc

import (
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/asdine/lobby/validation"
	"golang.org/x/net/context"
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

// Create a bucket in the registry.
func (s *registryService) Create(ctx context.Context, newBucket *proto.NewBucket) (*proto.Empty, error) {
	err := validation.Validate(newBucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	err = s.registry.Create(newBucket.Backend, newBucket.Name)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	return new(proto.Empty), nil
}

// Exists check a bucket in the registry.
func (s *registryService) Status(ctx context.Context, bucket *proto.Bucket) (*proto.BucketStatus, error) {
	err := validation.Validate(bucket)
	if err != nil {
		return nil, newError(err, s.logger)
	}

	var exists bool

	_, err = s.registry.Bucket(bucket.Name)
	if err != nil {
		if err != lobby.ErrBucketNotFound {
			return nil, newError(err, s.logger)
		}
	} else {
		exists = true
	}

	return &proto.BucketStatus{
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

// Create a bucket and register it to the Registry.
func (s *Registry) Create(backendName, bucketName string) error {
	_, err := s.client.Create(context.Background(), &proto.NewBucket{Name: bucketName, Backend: backendName})
	if err != nil {
		return errFromGRPC(err)
	}

	return nil
}

// Bucket returns the bucket associated with the given id.
func (s *Registry) Bucket(name string) (lobby.Bucket, error) {
	status, err := s.client.Status(context.Background(), &proto.Bucket{Name: name})
	if err != nil {
		return nil, errFromGRPC(err)
	}

	if !status.Exists {
		return nil, lobby.ErrBucketNotFound
	}

	return s.Backend.Bucket(name)
}

// Close the connexion to the Registry.
func (s *Registry) Close() error {
	return s.conn.Close()
}
