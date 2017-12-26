package app

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/asdine/lobby/log"
	"github.com/stretchr/testify/require"
)

func appHelper(t *testing.T) (*App, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	var app App
	app.out = ioutil.Discard
	app.Logger = log.New(log.Output(ioutil.Discard))
	app.errc = make(chan error)
	app.Config.Paths.DataDir = path.Join(dir, "data")
	app.Config.Paths.SocketDir = path.Join(app.Config.Paths.DataDir, "sockets")
	err = app.Config.Paths.Create()
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
