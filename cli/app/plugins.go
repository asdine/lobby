package app

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc"
	"github.com/pkg/errors"
)

func newBackendPluginsStep() *backendPluginsStep {
	return &backendPluginsStep{
		pluginLoader: rpc.LoadBackendPlugin,
	}
}

type backendPluginsStep struct {
	pluginLoader func(context.Context, string, string, string, string) (lobby.Backend, lobby.Plugin, error)
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
			app.ConfigPath,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to run backend '%s'", name)
		}

		app.Logger.Debugf("Started %s plugin \n", name)
		app.registry.RegisterBackend(name, bck)
		s.plugins = append(s.plugins, plg)

		app.wg.Add(1)
		go func(p lobby.Plugin) {
			defer app.wg.Done()

			err := p.Wait()
			if err != nil {
				app.Logger.Println(err)
				app.errc <- err
			}
		}(plg)
	}

	return nil
}

func (s *backendPluginsStep) teardown(ctx context.Context, app *App) error {
	for _, p := range s.plugins {
		err := p.Close()
		if err != nil {
			app.Logger.Printf("Error while closing plugin %s: %s\n", p.Name(), err)
			app.errc <- err
		}

		err = p.Wait()
		if err != nil {
			app.Logger.Printf("Error while waiting for plugin %s to close: %s\n", p.Name(), err)
			app.errc <- err
		}

		app.Logger.Debugf("Stopped %s plugin\n", p.Name())
	}

	return nil
}
