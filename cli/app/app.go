package app

import (
	"context"
	"io"
	"os"
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
	out      io.Writer
	registry lobby.Registry
	steps    steps
}

// Run all the app components. Can be gracefully shutdown using the provided context.
func (a *App) Run(ctx context.Context) error {
	var errs Errors
	a.errc = make(chan error)
	if a.out == nil {
		a.out = os.Stderr
	}

	a.Logger = log.New(
		log.Prefix("lobby:"),
		log.Output(a.out),
		log.Debug(a.Config.Debug),
	)

	a.logLobbyInfos()

	if a.steps == nil {
		a.steps = []step{
			directoriesStep(),
			new(registryStep),
			boltBackendStep(),
			newBackendPluginsStep(),
			newGRPCUnixSocketStep(a),
			newGRPCPortStep(a),
			newHTTPStep(a),
		}
	}

	err := a.steps.setup(ctx, a)
	if err != nil && err != context.Canceled {
		a.Logger.Println(err)
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
