package app

import (
	"context"
	"sync"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/log"
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

// Run all the app components. Can be gracefully shutdown using the provided context.
func (a *App) Run(ctx context.Context) error {
	var errs Errors
	a.errc = make(chan error)
	a.Logger = log.New(log.Prefix("lobby:"), log.Debug(a.Config.Debug))

	a.logLobbyInfos()

	if a.steps == nil {
		a.steps = []step{
			new(directoriesStep),
			new(registryStep),
			new(boltBackendStep),
			newBackendPluginsStep(),
			newGRPCUnixSocketStep(a),
			newGRPCPortStep(a),
			newHTTPStep(a),
		}
	}

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

func (a *App) logLobbyInfos() {
	a.Logger.Println("lobby Version:", lobby.Version)
	a.Logger.Debug("Debug mode enabled")
}
