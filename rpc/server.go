package rpc

import (
	"net"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"google.golang.org/grpc"
)

// NewServer returns a configured gRPC server.
func NewServer(services ...func(*grpc.Server)) lobby.Server {
	g := grpc.NewServer()

	for _, s := range services {
		s(g)
	}

	return &server{srv: g}
}

// WithTopicService enables the TopicService.
func WithTopicService(b lobby.Backend) func(*grpc.Server) {
	return func(g *grpc.Server) {
		proto.RegisterTopicServiceServer(g, newTopicService(b))
	}
}

// WithRegistryService enables the RegistryService.
func WithRegistryService(r lobby.Registry) func(*grpc.Server) {
	return func(g *grpc.Server) {
		proto.RegisterRegistryServiceServer(g, newRegistryService(r))
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
