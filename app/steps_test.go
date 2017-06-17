package app

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func appHelper(t *testing.T) (*App, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	var app App
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
