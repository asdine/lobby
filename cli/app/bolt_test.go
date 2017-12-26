package app

import (
	"context"
	"testing"

	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/require"
)

func TestBoltBackendStep(t *testing.T) {
	test := func(app *App) {
		reg := new(mock.Registry)
		app.registry = reg
		step := boltBackendStep()
		err := step.setup(context.Background(), app)
		require.NoError(t, err)
		bck, ok := reg.Backends["bolt"]
		require.True(t, ok)
		err = bck.Close()
		require.NoError(t, err)
	}

	t.Run("AsDefaultBackend", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		test(app)
	})

	t.Run("ExplicitBackend", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		app.Config.Plugins.Backends = []string{"a", "b"}
		app.Config.Bolt.Backend = true
		test(app)
	})

	t.Run("WithBackends", func(t *testing.T) {
		app, cleanup := appHelper(t)
		defer cleanup()

		reg := new(mock.Registry)
		app.registry = reg
		app.Config.Plugins.Backends = []string{"a", "b"}
		step := boltBackendStep()
		err := step.setup(context.Background(), app)
		require.NoError(t, err)
		_, ok := reg.Backends["bolt"]
		require.False(t, ok)
	})
}
