package app

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStep struct {
	setupFn    func(ctx context.Context, app *App) error
	teardownFn func(ctx context.Context, app *App) error
}

func (s *mockStep) setup(ctx context.Context, app *App) error {
	if s.setupFn != nil {
		return s.setupFn(ctx, app)
	}

	return nil
}

func (s *mockStep) teardown(ctx context.Context, app *App) error {
	if s.teardownFn != nil {
		return s.teardownFn(ctx, app)
	}

	return nil
}

func TestApp(t *testing.T) {
	t.Run("SetupError", func(t *testing.T) {
		var app App
		app.out = ioutil.Discard

		m := mockStep{
			setupFn: func(ctx context.Context, app *App) error {
				return errors.New("setup error")
			},
			teardownFn: func(ctx context.Context, app *App) error {
				return errors.New("teardown error")
			},
		}

		app.steps = []step{&m}

		err := app.Run(context.Background())
		errs := err.(Errors)
		assert.Len(t, errs, 2)
	})

	t.Run("Goroutine error", func(t *testing.T) {
		var app App
		app.out = ioutil.Discard

		m := mockStep{
			setupFn: func(ctx context.Context, app *App) error {
				app.wg.Add(1)
				go func() {
					defer app.wg.Done()

					app.errc <- errors.New("goroutine error")
				}()
				return nil
			},
			teardownFn: func(ctx context.Context, app *App) error {
				return nil
			},
		}

		app.steps = []step{&m}

		err := app.Run(context.Background())
		errs := err.(Errors)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "goroutine error")
	})

	t.Run("Cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var app App
		app.out = ioutil.Discard
		quit := make(chan struct{})

		m := mockStep{
			setupFn: func(ctx context.Context, app *App) error {
				app.wg.Add(1)
				go func() {
					defer app.wg.Done()

					<-quit
				}()

				cancel()
				return nil
			},
			teardownFn: func(ctx context.Context, app *App) error {
				quit <- struct{}{}

				return nil
			},
		}

		app.steps = []step{&m}

		err := app.Run(ctx)
		require.NoError(t, err)
	})
}
