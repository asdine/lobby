package app

import (
	"context"
	"path"

	"github.com/asdine/lobby/bolt"
)

func boltBackendStep() step {
	return setupFunc(func(ctx context.Context, app *App) error {
		if len(app.Config.Plugins.Backends) > 0 && !app.Config.Bolt.Backend {
			return nil
		}

		dataPath := path.Join(app.Config.Paths.DataDir, "db")
		err := createDir(dataPath)
		if err != nil {
			return err
		}

		boltPath := path.Join(dataPath, "bolt")
		err = createDir(boltPath)
		if err != nil {
			return err
		}

		backendPath := path.Join(boltPath, "backend.db")

		// Creating default backend.
		bck, err := bolt.NewBackend(backendPath)
		if err != nil {
			return err
		}

		app.registry.RegisterBackend("bolt", bck)
		return nil
	})
}
