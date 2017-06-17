package app

import (
	"context"
	"log"
	"net"
	"os"
	"path"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/rpc"
)

type step interface {
	setup(context.Context, *App) error
	teardown(context.Context, *App) error
}

type steps []step

func (s steps) setup(ctx context.Context, app *App) error {
	for _, step := range s {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := step.setup(ctx, app)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s steps) teardown(ctx context.Context, app *App) []error {
	var errs []error

	for _, step := range s {
		err := step.teardown(ctx, app)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type directoriesStep int

func (directoriesStep) setup(ctx context.Context, app *App) error {
	return app.Options.Paths.Create()
}

func (directoriesStep) teardown(ctx context.Context, app *App) error {
	return nil
}

type registryStep int

func (registryStep) setup(ctx context.Context, app *App) error {
	dataPath := path.Join(app.Options.Paths.ConfigDir, "data")
	err := createDir(dataPath)
	if err != nil {
		return err
	}

	boltPath := path.Join(dataPath, "bolt")
	err = createDir(boltPath)
	if err != nil {
		return err
	}

	registryPath := path.Join(boltPath, "registry.db")
	backendPath := path.Join(boltPath, "backend.db")

	// Creating default registry.
	reg, err := bolt.NewRegistry(registryPath)
	if err != nil {
		return err
	}
	app.registry = reg

	// Creating default backend.
	bck, err := bolt.NewBackend(backendPath)
	if err != nil {
		return err
	}

	reg.RegisterBackend("bolt", bck)
	return nil
}

func (registryStep) teardown(ctx context.Context, app *App) error {
	if app.registry != nil {
		err := app.registry.Close()
		app.registry = nil
		return err
	}

	return nil
}

func newGRPCUnixSocketStep() *gRPCUnixSocketStep {
	return &gRPCUnixSocketStep{
		serverStep: &serverStep{
			logger: log.New(os.Stderr, "[gRPC] ", log.LstdFlags),
		},
	}
}

type gRPCUnixSocketStep struct {
	*serverStep
}

func (g *gRPCUnixSocketStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("unix", path.Join(app.Options.Paths.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		rpc.WithBucketService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return g.runServer(srv, l, app)
}

func newGRPCPortStep() *gRPCPortStep {
	return &gRPCPortStep{
		serverStep: &serverStep{
			logger: log.New(os.Stderr, "[gRPC] ", log.LstdFlags),
		},
	}
}

type gRPCPortStep struct {
	*serverStep
}

func (g *gRPCPortStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("tcp", ":5656")
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		rpc.WithBucketService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return g.runServer(srv, l, app)
}

func newHttpStep() *httpStep {
	return &httpStep{
		serverStep: &serverStep{
			logger: log.New(os.Stderr, "[http] ", log.LstdFlags),
		},
	}
}

type httpStep struct {
	*serverStep
}

func (h *httpStep) setup(ctx context.Context, app *App) error {
	l, err := net.Listen("tcp", ":5657")
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		rpc.WithBucketService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return h.runServer(srv, l, app)
}

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
		s.logger.Printf("Listening %s requests on %s.\n", srv.Name(), l.Addr().String())
		close(c)
		_ = srv.Serve(l)
	}()

	<-c
	return nil
}

func (s *serverStep) teardown(ctx context.Context, app *App) error {
	err := s.srv.Stop()
	s.srv = nil
	return err
}
