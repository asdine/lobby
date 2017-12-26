package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendPluginsSteps(t *testing.T) {
	t.Run("ErrorsDuringSetup", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Config.Paths.DataDir = "dataDir"
		app.Config.Paths.PluginDir = "pluginDir"
		app.Config.Plugins.Backends = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Config.Plugins.Backends[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		var i int
		s.pluginLoader = func(ctx context.Context, name, cmdPath, dataDir, configFile string) (lobby.Backend, lobby.Plugin, error) {
			i++
			if i == 3 {
				return nil, nil, errors.New("unexpected error")
			}

			return new(mock.Backend), new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.Error(t, err)
		require.Len(t, s.plugins, 2)
		require.Len(t, m.Backends, 2)

		err = s.teardown(context.Background(), app)
		require.NoError(t, err)
		for _, p := range s.plugins {
			require.Equal(t, 1, p.(*mock.Plugin).CloseInvoked)
		}
	})

	t.Run("ErrorsDuringTeardown", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Config.Paths.DataDir = "dataDir"
		app.Config.Paths.PluginDir = "pluginDir"
		app.Config.Plugins.Backends = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Config.Plugins.Backends[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		s.pluginLoader = func(ctx context.Context, name, cmdPath, dataDir, configFile string) (lobby.Backend, lobby.Plugin, error) {
			return new(mock.Backend), new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.NoError(t, err)
		require.Len(t, s.plugins, 5)
		require.Len(t, m.Backends, 5)

		s.plugins[3].(*mock.Plugin).CloseFn = func() error {
			return errors.New("unexpected error")
		}

		c := make(chan struct{})

		go func() {
			err = s.teardown(context.Background(), app)
			assert.NoError(t, err)
			close(c)
		}()

		require.EqualError(t, <-app.errc, "unexpected error")
		<-c
		for i, p := range s.plugins {
			if i != 3 {
				require.Equal(t, 1, p.(*mock.Plugin).CloseInvoked)
			}
		}
	})

	t.Run("OK", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Config.Paths.DataDir = "dataDir"
		app.Config.Paths.PluginDir = "pluginDir"
		app.Config.Plugins.Backends = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Config.Plugins.Backends[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		var i int
		s.pluginLoader = func(ctx context.Context, name, cmdPath, dataDir, configFile string) (lobby.Backend, lobby.Plugin, error) {
			require.Equal(t, fmt.Sprintf("plugin%d", i), name)
			require.Equal(t, fmt.Sprintf("pluginDir/lobby-plugin%d", i), cmdPath)
			require.Equal(t, "dataDir", dataDir)
			i++
			return new(mock.Backend), new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.NoError(t, err)
		require.Len(t, s.plugins, 5)
		require.Len(t, m.Backends, 5)

		err = s.teardown(context.Background(), app)
		require.NoError(t, err)
		for _, p := range s.plugins {
			require.Equal(t, 1, p.(*mock.Plugin).CloseInvoked)
		}
	})
}
