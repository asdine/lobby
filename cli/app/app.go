package app

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/asdine/lobby"
)

// App is the main application. It bootstraps all the components
// and can be gracefully shutdown.
type App struct {
	Config Config
	Logger *log.Logger

	wg       sync.WaitGroup
	errc     chan error
	registry lobby.Registry
	steps    steps
}

// NewApp create a configured App.
func NewApp() *App {
	app := App{
		Logger: log.New(os.Stderr, "[lobby] ", log.LstdFlags),
		errc:   make(chan error),
	}

	app.steps = []step{
		new(directoriesStep),
		new(registryStep),
		newGRPCUnixSocketStep(),
		newGRPCPortStep(),
		newHTTPStep(),
		newBackendPluginsStep(),
		newServerPluginsStep(),
	}

	return &app
}

// Run all the app components. Can be gracefully shutdown using the provided context.
func (a *App) Run(ctx context.Context) error {
	var errs Errors

	err := a.steps.setup(ctx, a)
	if err != nil && err != context.Canceled {
		errs = append(errs, err)
	}

	if err == nil {
		// block until either an error or a cancel happens
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

	closeErrs := a.steps.teardown(ctx, a)
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
