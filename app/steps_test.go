package app

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appHelper(t *testing.T) (*App, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	var app App
	app.errc = make(chan error)
	app.Options.Paths.ConfigDir = path.Join(dir, "config")
	app.Options.Paths.SocketDir = path.Join(app.Options.Paths.ConfigDir, "sockets")
	err = app.Options.Paths.Create()
	require.NoError(t, err)

	return &app, func() {
		os.RemoveAll(dir)
	}
}

func TestSteps(t *testing.T) {
	t.Run("Cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		s := steps([]step{
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					cancel()
					return nil
				},
			},
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					t.Error("Should not be called")
					return nil
				},
			},
		})

		err := s.setup(ctx, nil)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("ReturnOnError", func(t *testing.T) {
		s := steps([]step{
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					return nil
				},
			},
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					return errors.New("unexpected error")
				},
			},
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					t.Error("Should not be called")
					return nil
				},
			},
		})

		err := s.setup(context.Background(), nil)
		require.EqualError(t, err, "unexpected error")
	})

	t.Run("ErrorsOnTeardown", func(t *testing.T) {
		s := steps([]step{
			&mockStep{
				teardownFn: func(ctx context.Context, app *App) error {
					return errors.New("3")
				},
			},
			&mockStep{
				teardownFn: func(ctx context.Context, app *App) error {
					return errors.New("2")
				},
			},
			&mockStep{
				teardownFn: func(ctx context.Context, app *App) error {
					return errors.New("1")
				},
			},
		})

		err := s.setup(context.Background(), nil)
		require.NoError(t, err)
		errs := s.teardown(context.Background(), nil)
		require.Len(t, errs, 3)
		require.EqualError(t, errs[0], "1")
		require.EqualError(t, errs[1], "2")
		require.EqualError(t, errs[2], "3")
	})

	t.Run("OK", func(t *testing.T) {
		s := steps([]step{
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					return nil
				},
			},
			&mockStep{
				setupFn: func(ctx context.Context, app *App) error {
					return nil
				},
			},
		})

		err := s.setup(context.Background(), nil)
		require.NoError(t, err)
	})
}

func TestRegistryStep(t *testing.T) {
	app, cleanup := appHelper(t)
	defer cleanup()

	var r registryStep
	err := r.setup(context.Background(), app)
	require.NoError(t, err)
	require.NotNil(t, app.registry)

	err = r.teardown(context.Background(), app)
	require.NoError(t, err)
	require.Nil(t, app.registry)
}

func TestServersSteps(t *testing.T) {
	app, cleanup := appHelper(t)
	defer cleanup()

	testCases := []step{
		newGRPCUnixSocketStep(),
		newGRPCPortStep(),
		newHTTPStep(),
	}

	for _, s := range testCases {
		err := s.setup(context.Background(), app)
		require.NoError(t, err)

		err = s.teardown(context.Background(), app)
		require.NoError(t, err)
	}

	app.wg.Wait()
}

func TestServerPluginsSteps(t *testing.T) {
	t.Run("ErrorsDuringSetup", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Server = make([]string, 5)
		for i := 0; i < 5; i++ {
			app.Options.Plugins.Server[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newServerPluginsStep()
		var i int
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Plugin, error) {
			i++
			if i == 3 {
				return nil, errors.New("unexpected error")
			}

			return new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.Error(t, err)
		require.Len(t, s.plugins, 2)

		err = s.teardown(context.Background(), app)
		require.NoError(t, err)
		for _, p := range s.plugins {
			require.Equal(t, 1, p.(*mock.Plugin).CloseInvoked)
		}
	})

	t.Run("ErrorsDuringTeardown", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Server = make([]string, 5)
		for i := 0; i < 5; i++ {
			app.Options.Plugins.Server[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newServerPluginsStep()
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Plugin, error) {
			return new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.NoError(t, err)
		require.Len(t, s.plugins, 5)

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

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Server = make([]string, 5)
		for i := 0; i < 5; i++ {
			app.Options.Plugins.Server[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newServerPluginsStep()
		var i int
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Plugin, error) {
			require.Equal(t, fmt.Sprintf("plugin%d", i), name)
			require.Equal(t, fmt.Sprintf("pluginDir/lobby-plugin%d", i), cmdPath)
			require.Equal(t, "configDir", configDir)
			i++
			return new(mock.Plugin), nil
		}

		err := s.setup(context.Background(), app)
		require.NoError(t, err)
		require.Len(t, s.plugins, 5)

		err = s.teardown(context.Background(), app)
		require.NoError(t, err)
		for _, p := range s.plugins {
			require.Equal(t, 1, p.(*mock.Plugin).CloseInvoked)
		}
	})
}

func TestBackendPluginsSteps(t *testing.T) {
	t.Run("ErrorsDuringSetup", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Backend = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Options.Plugins.Backend[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		var i int
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Backend, lobby.Plugin, error) {
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

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Backend = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Options.Plugins.Backend[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Backend, lobby.Plugin, error) {
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

		app.Options.Paths.ConfigDir = "configDir"
		app.Options.Paths.PluginDir = "pluginDir"
		app.Options.Plugins.Backend = make([]string, 5)
		var m mock.Registry
		app.registry = &m

		for i := 0; i < 5; i++ {
			app.Options.Plugins.Backend[i] = fmt.Sprintf("plugin%d", i)
		}

		s := newBackendPluginsStep()
		var i int
		s.pluginLoader = func(name, cmdPath, configDir string) (lobby.Backend, lobby.Plugin, error) {
			require.Equal(t, fmt.Sprintf("plugin%d", i), name)
			require.Equal(t, fmt.Sprintf("pluginDir/lobby-plugin%d", i), cmdPath)
			require.Equal(t, "configDir", configDir)
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
