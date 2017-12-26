package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServersSteps(t *testing.T) {
	app, cleanup := appHelper(t)
	defer cleanup()

	testCases := []step{
		newHTTPStep(app),
		newGRPCUnixSocketStep(app),
		newGRPCPortStep(app),
	}

	for _, s := range testCases {
		err := s.setup(context.Background(), app)
		require.NoError(t, err)
	}

	// bug when calling stop right after serve on http.
	time.Sleep(10 * time.Millisecond)

	for _, s := range testCases {
		err := s.teardown(context.Background(), app)
		require.NoError(t, err)
	}

	app.wg.Wait()
}
