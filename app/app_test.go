package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	app := NewApp()

	app.bootstrap = func(ctx context.Context, app *App) {
		for i := 0; i < 10; i++ {
			app.errc <- fmt.Errorf("error %d", i)
		}

		close(app.errc)
	}

	err := app.Run(context.Background())
	errs := err.(Errors)
	require.Len(t, errs, 10)
}

func testBootstrap(ctx context.Context, t *testing.T, app *App, b *bootstrapper) []error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		b.bootstrap(ctx, app)
	}()

	var errs []error
	for err := range b.errc {
		errs = append(errs, err)
	}

	wg.Wait()
	return errs
}

func TestBootstrap(t *testing.T) {
	t.Run("Cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		b := newBootstrapper([]step{
			func(ctx context.Context, app *App) error {
				cancel()
				return nil
			},
			func(ctx context.Context, app *App) error {
				t.Error("Should not be called")
				return nil
			},
		})

		errs := testBootstrap(ctx, t, nil, b)
		require.Len(t, errs, 1)
		require.Equal(t, context.Canceled, errs[0])
	})

	t.Run("ReturnOnError", func(t *testing.T) {
		b := newBootstrapper([]step{
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

		errs := testBootstrap(context.Background(), t, nil, b)
		require.Len(t, errs, 1)
		require.EqualError(t, errs[0], "unexpected error")
	})

	t.Run("OK", func(t *testing.T) {
		b := newBootstrapper([]step{
			func(ctx context.Context, app *App) error {
				return nil
			},
			func(ctx context.Context, app *App) error {
				return nil
			},
		})
		errs := testBootstrap(context.Background(), t, nil, b)
		require.Nil(t, errs)
	})
}
