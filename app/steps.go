package app

import (
	"context"
	"path"

	"github.com/asdine/lobby/bolt"
)

type step interface {
	setup(context.Context, *App) error
	teardown(context.Context, *App) error
}

type steps []step

func (s steps) setup(ctx context.Context, app *App) error {
	for _, step := range s {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := step.setup(ctx, app)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s steps) teardown(ctx context.Context, app *App) []error {
	var errs []error

	for _, step := range s {
		err := step.teardown(ctx, app)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type directoriesStep int

func (directoriesStep) setup(ctx context.Context, app *App) error {
	return app.Options.Paths.Create()
}

func (directoriesStep) teardown(ctx context.Context, app *App) error {
	return nil
}

type registryStep int

func (registryStep) setup(ctx context.Context, app *App) error {
	dataPath := path.Join(app.Options.Paths.ConfigDir, "data")
	err := createDir(dataPath)
	if err != nil {
		return err
	}

	boltPath := path.Join(dataPath, "bolt")
	err = createDir(boltPath)
	if err != nil {
		return err
	}

	registryPath := path.Join(boltPath, "registry.db")
	backendPath := path.Join(boltPath, "backend.db")

	// Creating default registry.
	reg, err := bolt.NewRegistry(registryPath)
	if err != nil {
		return err
	}
	app.registry = reg

	// Creating default backend.
	bck, err := bolt.NewBackend(backendPath)
	if err != nil {
		return err
	}

	reg.RegisterBackend("bolt", bck)
	return nil
}

func (registryStep) teardown(ctx context.Context, app *App) error {
	if app.registry != nil {
		err := app.registry.Close()
		app.registry = nil
		return err
	}

	return nil
}
