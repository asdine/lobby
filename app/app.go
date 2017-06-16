package app

import (
	"context"
	"log"
	"os"
)

type Options struct {
	Paths Paths
}

type App struct {
	Options Options
	Logger  *log.Logger
	*bootstrapper
}

func NewApp() *App {
	app := App{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}

	app.bootstrapper = newBootstrapper([]step{
		stepCreateDir,
	})

	return &app
}

func (a *App) Run(ctx context.Context) error {
	var errs Errors

	go a.bootstrap(ctx, a)

	for err := range a.errc {
		errs = append(errs, err)
	}

	return errs
}

func newBootstrapper(steps []step) *bootstrapper {
	b := bootstrapper{
		steps: steps,
		errc:  make(chan error),
	}

	b.bootstrap = func(ctx context.Context, app *App) {
		bootstrap(ctx, &b, app)
	}

	return &b
}

type bootstrapper struct {
	steps     []step
	errc      chan error
	bootstrap func(context.Context, *App)
	rollback  func(context.Context, *App, chan error)
}

func bootstrap(ctx context.Context, b *bootstrapper, app *App) {
	defer close(b.errc)

	for _, fn := range b.steps {
		select {
		case <-ctx.Done():
			b.errc <- ctx.Err()
			if b.rollback != nil {
				b.rollback(ctx, app, b.errc)
			}
			return
		default:
			err := fn(ctx, app)
			if err != nil {
				b.errc <- err
				if b.rollback != nil {
					b.rollback(ctx, app, b.errc)
				}
				return
			}
		}
	}
}

type step func(context.Context, *App) error

func stepCreateDir(ctx context.Context, app *App) error {
	return app.Options.Paths.Create()
}
