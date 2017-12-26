package app

import (
	"context"
	"fmt"
	"net"
	"path"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/http"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/rpc"
)

type serverStep struct {
	logger *log.Logger
	srv    lobby.Server
}

func (s *serverStep) runServer(srv lobby.Server, l net.Listener, app *App) error {
	c := make(chan struct{})

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		s.srv = srv
		s.logger.Printf("Listening for requests on %s.\n", l.Addr().String())
		close(c)
		err := srv.Serve(l)
		if err != nil {
			s.logger.Println(err)
			app.errc <- err
		}
	}()

	<-c
	return nil
}

func (s *serverStep) teardown(ctx context.Context, app *App) error {
	s.logger.Debugf("Shutting down")
	if s.srv != nil {
		err := s.srv.Stop()
		s.srv = nil
		return err
	}

	return nil
}

func newGRPCUnixSocketStep(app *App) *gRPCUnixSocketStep {
	return &gRPCUnixSocketStep{
		serverStep: &serverStep{
			logger: log.New(
				log.Prefix("gRPC server:"),
				log.Output(app.out),
				log.Debug(app.Config.Debug),
			),
		},
	}
}

type gRPCUnixSocketStep struct {
	*serverStep
}

func (g *gRPCUnixSocketStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("unix", path.Join(app.Config.Paths.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		g.serverStep.logger,
		rpc.WithTopicService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return g.runServer(srv, l, app)
}

func newGRPCPortStep(app *App) *gRPCPortStep {
	return &gRPCPortStep{
		serverStep: &serverStep{
			logger: log.New(
				log.Prefix("gRPC server:"),
				log.Output(app.out),
				log.Debug(app.Config.Debug),
			),
		},
	}
}

type gRPCPortStep struct {
	*serverStep
}

func (g *gRPCPortStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.Grpc.Port))
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		g.serverStep.logger,
		rpc.WithTopicService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return g.runServer(srv, l, app)
}

func newHTTPStep(app *App) *httpStep {
	return &httpStep{
		serverStep: &serverStep{
			logger: log.New(
				log.Prefix("http server:"),
				log.Output(app.out),
				log.Debug(app.Config.Debug),
			),
		},
	}
}

type httpStep struct {
	*serverStep
}

func (h *httpStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.HTTP.Port))
	if err != nil {
		return err
	}

	srv := http.NewServer(
		http.NewHandler(app.registry, h.logger),
	)
	return h.runServer(srv, l, app)
}
