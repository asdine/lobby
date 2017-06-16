package app

import (
	"context"
	"log"
	"os"
	"sync"
)

type Options struct {
	Paths Paths
}

type App struct {
	Options  Options
	Logger   *log.Logger
	setup    *setup
	teardown *teardown
	wg       sync.WaitGroup
	errc     chan error
}

func NewApp() *App {
	app := App{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		errc:   make(chan error),
	}

	setupFns := []step{
		stepCreateDir,
	}

	app.setup = newSetup(setupFns)
	app.teardown = newTeardown(nil)

	return &app
}

func (a *App) Run(ctx context.Context) error {
	var errs Errors

	err := a.setup.setup(ctx, a)
	if err != nil && err != context.Canceled {
		errs = append(errs, err)
	}

	if err == nil {
		// block until either an error or a cancel
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != context.Canceled {
				errs = append(errs, err)
			}
		case err := <-a.errc:
			errs = append(errs, err)
		}
	}

	errsC := make(chan Errors)
	defer close(errsC)

	// get errors from any goroutine
	go func() {
		var errs Errors
		for err := range a.errc {
			errs = append(errs, err)
		}
		errsC <- errs
	}()

	closeErrs := a.teardown.teardown(ctx, a)
	if len(closeErrs) != 0 {
		errs = append(errs, closeErrs...)
	}

	a.wg.Wait()
	close(a.errc)
	errs = append(errs, <-errsC...)

	if len(errs) != 0 {
		return errs
	}
	return nil
}

type step func(context.Context, *App) error

func stepCreateDir(ctx context.Context, app *App) error {
	return app.Options.Paths.Create()
}

func newSetup(steps []step) *setup {
	return &setup{
		steps: steps,
	}
}

type setup struct {
	steps []step
}

func (s *setup) setup(ctx context.Context, app *App) error {
	for _, fn := range s.steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := fn(ctx, app)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type teardown struct {
	steps []step
}

func newTeardown(steps []step) *teardown {
	return &teardown{
		steps: steps,
	}
}

func (t *teardown) teardown(ctx context.Context, app *App) []error {
	var errs []error

	for _, fn := range t.steps {
		err := fn(ctx, app)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
