package rpc

import (
	"net"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc/proto"
	"google.golang.org/grpc"
)

// NewServer returns a configured gRPC server.
func NewServer(r lobby.Registry) lobby.Server {
	g := grpc.NewServer()
	bs := newBucketService(r)
	rs := newRegistryService(r)

	proto.RegisterBucketServiceServer(g, bs)
	proto.RegisterRegistryServiceServer(g, rs)

	return &server{srv: g}
}

type server struct {
	srv *grpc.Server
}

func (s *server) Serve(l net.Listener) error {
	return s.srv.Serve(l)
}

func (s *server) Stop() error {
	s.srv.GracefulStop()
	return nil
}
