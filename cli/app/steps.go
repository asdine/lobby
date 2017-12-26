package app

import (
	"context"
	"fmt"
	"net"
	"path"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/etcd"
	"github.com/asdine/lobby/http"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/rpc"
	"github.com/coreos/etcd/clientv3"
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

func setupFunc(fn func(ctx context.Context, app *App) error) step {
	return &stepFn{fn: fn}
}

type stepFn struct {
	fn func(ctx context.Context, app *App) error
}

func (s *stepFn) setup(ctx context.Context, app *App) error {
	return s.fn(ctx, app)
}

func (s *stepFn) teardown(ctx context.Context, app *App) error {
	return nil
}

func directoriesStep() step {
	return setupFunc(func(ctx context.Context, app *App) error {
		return app.Config.Paths.Create()
	})
}

type registryStep int

func (registryStep) setup(ctx context.Context, app *App) error {
	var reg lobby.Registry
	var err error
	switch app.Config.Registry {
	case "":
		fallthrough
	case "bolt":
		app.Logger.Debug("Using bolt registry")
		reg, err = boltRegistry(ctx, app)
	case "etcd":
		app.Logger.Debug("Using etcd registry")
		reg, err = etcdRegistry(ctx, app)
	default:
		err = errors.New("unknown registry")
	}
	if err != nil {
		return err
	}

	app.registry = reg
	return nil
}

func boltRegistry(ctx context.Context, app *App) (lobby.Registry, error) {
	dataPath := path.Join(app.Config.Paths.DataDir, "db")
	err := createDir(dataPath)
	if err != nil {
		return nil, err
	}

	boltPath := path.Join(dataPath, "bolt")
	err = createDir(boltPath)
	if err != nil {
		return nil, err
	}

	registryPath := path.Join(boltPath, "registry.db")

	return bolt.NewRegistry(registryPath, log.New(log.Prefix("bolt registry:"), log.Debug(app.Config.Debug)))
}

func etcdRegistry(ctx context.Context, app *App) (lobby.Registry, error) {
	client, err := clientv3.New(app.Config.Etcd)
	if err != nil {
		return nil, err
	}

	return etcd.NewRegistry(
		client,
		log.New(log.Prefix("etcd registry:"), log.Debug(app.Config.Debug)),
		"lobby",
	)
}

func (registryStep) teardown(ctx context.Context, app *App) error {
	if app.registry != nil {
		app.Logger.Debug("Closing registry")
		err := app.registry.Close()
		app.registry = nil
		return err
	}

	return nil
}

func boltBackendStep() step {
	return setupFunc(func(ctx context.Context, app *App) error {
		if len(app.Config.Plugins.Backends) > 0 && !app.Config.Bolt.Backend {
			return nil
		}

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

		backendPath := path.Join(boltPath, "backend.db")

		// Creating default backend.
		bck, err := bolt.NewBackend(backendPath)
		if err != nil {
			return err
		}

		app.registry.RegisterBackend("bolt", bck)
		return nil
	})
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

func newBackendPluginsStep() *backendPluginsStep {
	return &backendPluginsStep{
		pluginLoader: rpc.LoadBackendPlugin,
	}
}

type backendPluginsStep struct {
	pluginLoader func(context.Context, string, string, string) (lobby.Backend, lobby.Plugin, error)
	plugins      []lobby.Plugin
}

func (s *backendPluginsStep) setup(ctx context.Context, app *App) error {
	for _, name := range app.Config.Plugins.Backends {
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

		app.Logger.Debugf("Started %s plugin \n", name)
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

		err = p.Wait()
		if err != nil {
			app.errc <- err
		}

		app.Logger.Debugf("Stopped %s plugin\n", p.Name())
	}

	return nil
}
