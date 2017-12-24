package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/http"
	"github.com/asdine/lobby/rpc"
	"github.com/pkg/errors"
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

	for i := len(s) - 1; i >= 0; i-- {
		err := s[i].teardown(ctx, app)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type directoriesStep int

func (directoriesStep) setup(ctx context.Context, app *App) error {
	return app.Config.Paths.Create()
}

func (directoriesStep) teardown(ctx context.Context, app *App) error {
	return nil
}

type registryStep int

func (registryStep) setup(ctx context.Context, app *App) error {
	dataPath := path.Join(app.Config.Paths.DataDir, "db")
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
	l, err := net.Listen("unix", path.Join(app.Config.Paths.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}

	srv := rpc.NewServer(
		rpc.WithTopicService(app.registry),
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
		rpc.WithTopicService(app.registry),
		rpc.WithRegistryService(app.registry),
	)
	return g.runServer(srv, l, app)
}

func newHTTPStep() *httpStep {
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

	srv := http.NewServer(
		http.NewHandler(app.registry, h.logger),
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
		err := srv.Serve(l)
		if err != nil {
			s.logger.Println(err)
		}
	}()

	<-c
	return nil
}

func (s *serverStep) teardown(ctx context.Context, app *App) error {
	err := s.srv.Stop()
	s.srv = nil
	return err
}

func newBackendPluginsStep() *backendPluginsStep {
	return &backendPluginsStep{
		pluginLoader: rpc.LoadBackendPlugin,
		logger:       log.New(os.Stderr, "[lobby] ", log.LstdFlags),
	}
}

type backendPluginsStep struct {
	logger       *log.Logger
	pluginLoader func(context.Context, string, string, string) (lobby.Backend, lobby.Plugin, error)
	plugins      []lobby.Plugin
}

func (s *backendPluginsStep) setup(ctx context.Context, app *App) error {
	for _, name := range app.Config.Plugins.Backend {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		bck, plg, err := s.pluginLoader(
			ctx,
			name,
			path.Join(app.Config.Paths.PluginDir, fmt.Sprintf("lobby-%s", name)),
			app.Config.Paths.DataDir,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to run backend '%s'", name)
		}

		app.registry.RegisterBackend(name, bck)
		s.plugins = append(s.plugins, plg)
	}

	return nil
}

func (s *backendPluginsStep) teardown(ctx context.Context, app *App) error {
	for _, p := range s.plugins {
		err := p.Close()
		if err != nil {
			app.errc <- err
		}
	}

	return nil
}
