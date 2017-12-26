package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryStep(t *testing.T) {
	app, cleanup := appHelper(t)
	defer cleanup()

	var r registryStep
	err := r.setup(context.Background(), app)
	require.NoError(t, err)
	require.NotNil(t, app.registry)

	err = r.teardown(context.Background(), app)
	require.NoError(t, err)
	require.Nil(t, app.registry)
}
