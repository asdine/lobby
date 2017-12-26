package app

import (
	"context"
)

func directoriesStep() step {
	return setupFunc(func(ctx context.Context, app *App) error {
		return app.Config.Paths.Create()
	})
}
