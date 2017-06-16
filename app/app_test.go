package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	t.Run("SetupError", func(t *testing.T) {
		app := NewApp()

		app.setup.steps = []step{func(ctx context.Context, app *App) error {
			return errors.New("setup error")
		}}

		app.teardown.steps = []step{func(ctx context.Context, app *App) error {
			return errors.New("teardown error")
		}}

		err := app.Run(context.Background())
		errs := err.(Errors)
		assert.Len(t, errs, 2)
	})

	t.Run("Goroutine error", func(t *testing.T) {
		app := NewApp()

		app.setup.steps = []step{func(ctx context.Context, app *App) error {
			app.wg.Add(1)
			go func() {
				defer app.wg.Done()

				app.errc <- errors.New("goroutine error")
			}()
			return nil
		}}

		err := app.Run(context.Background())
		errs := err.(Errors)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "goroutine error")
	})

	t.Run("Cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		app := NewApp()
		quit := make(chan struct{})

		app.setup.steps = []step{func(ctx context.Context, app *App) error {
			app.wg.Add(1)
			go func() {
				defer app.wg.Done()

				<-quit
			}()

			cancel()
			return nil
		}}

		app.teardown.steps = []step{func(ctx context.Context, app *App) error {
			quit <- struct{}{}

			return nil
		}}

		err := app.Run(ctx)
		require.NoError(t, err)
	})
}

func TestBootstrap(t *testing.T) {
	t.Run("Cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		s := newSetup([]step{
			func(ctx context.Context, app *App) error {
				cancel()
				return nil
			},
			func(ctx context.Context, app *App) error {
				t.Error("Should not be called")
				return nil
			},
		})

		err := s.setup(ctx, nil)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("ReturnOnError", func(t *testing.T) {
		s := newSetup([]step{
			func(ctx context.Context, app *App) error {
				return nil
			},
			func(ctx context.Context, app *App) error {
				return errors.New("unexpected error")
			},
			func(ctx context.Context, app *App) error {
				t.Error("Should not be called")
				return nil
			},
		})

		err := s.setup(context.Background(), nil)
		require.EqualError(t, err, "unexpected error")
	})

	t.Run("OK", func(t *testing.T) {
		s := newSetup([]step{
			func(ctx context.Context, app *App) error {
				return nil
			},
			func(ctx context.Context, app *App) error {
				return nil
			},
		})

		err := s.setup(context.Background(), nil)
		require.NoError(t, err)
	})
}
