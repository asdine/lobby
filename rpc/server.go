package rpc

import (
	"net"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/rpc/proto"
	"google.golang.org/grpc"
)

// NewServer returns a configured gRPC server.
func NewServer(logger *log.Logger, services ...func(*grpc.Server, *log.Logger)) lobby.Server {
	g := grpc.NewServer()

	for _, s := range services {
		s(g, logger)
	}

	return &server{srv: g}
}

// WithTopicService enables the TopicService.
func WithTopicService(b lobby.Backend) func(*grpc.Server, *log.Logger) {
	return func(g *grpc.Server, logger *log.Logger) {
		proto.RegisterTopicServiceServer(g, newTopicService(b, logger))
	}
}

// WithRegistryService enables the RegistryService.
func WithRegistryService(r lobby.Registry) func(*grpc.Server, *log.Logger) {
	return func(g *grpc.Server, logger *log.Logger) {
		proto.RegisterRegistryServiceServer(g, newRegistryService(r, logger))
	}
}

type server struct {
	srv *grpc.Server
}

func (s *server) Name() string {
	return "gRPC"
}

func (s *server) Serve(l net.Listener) error {
	return s.srv.Serve(l)
}

func (s *server) Stop() error {
	s.srv.GracefulStop()
	return nil
}
