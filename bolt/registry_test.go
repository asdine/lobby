package bolt_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/require"
)

type errorHandler interface {
	Error(args ...interface{})
}

func preparePath(t errorHandler, dbName string) (string, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "lobby")
	if err != nil {
		t.Error(err)
	}

	return filepath.Join(dir, dbName), func() {
		os.RemoveAll(dir)
	}
}

func TestRegistry(t *testing.T) {
	_, cleanupStore := preparePath(t, "backend.db")
	defer cleanupStore()

	var b mock.Backend
	b.EndpointFn = func(path string) (lobby.Endpoint, error) {
		return new(mock.Endpoint), nil
	}
	defer b.Close()

	t.Run("create", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := bolt.NewRegistry(pathReg, log.New(log.Output(ioutil.Discard)))
		require.NoError(t, err)
		defer r.Close()

		_, err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrBackendNotFound, err)

		r.RegisterBackend("bolt1", &b)
		r.RegisterBackend("bolt2", &b)

		_, err = r.Create("bolt1", "a")
		require.NoError(t, err)

		_, err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrEndpointAlreadyExists, err)

		_, err = r.Create("bolt1", "b")
		require.NoError(t, err)

		_, err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrEndpointAlreadyExists, err)
	})

	t.Run("endpoint", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := bolt.NewRegistry(pathReg, log.New(log.Output(ioutil.Discard)))
		require.NoError(t, err)
		defer r.Close()

		r.RegisterBackend("bolt1", &b)
		r.RegisterBackend("bolt2", &b)

		b, err := r.Endpoint("")
		require.Equal(t, lobby.ErrEndpointNotFound, err)

		b, err = r.Endpoint("a")
		require.Equal(t, lobby.ErrEndpointNotFound, err)

		_, err = r.Create("bolt1", "a")
		require.NoError(t, err)

		b, err = r.Endpoint("a")
		require.NoError(t, err)
		require.NotNil(t, b)

		_, err = r.Create("bolt2", "b")
		require.NoError(t, err)

		b, err = r.Endpoint("b")
		require.NoError(t, err)
		require.NotNil(t, b)

		_, err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrEndpointAlreadyExists, err)
	})

	t.Run("endpoints", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := bolt.NewRegistry(pathReg, log.New(log.Output(ioutil.Discard)))
		require.NoError(t, err)
		defer r.Close()

		r.RegisterBackend("bolt1", &b)
		r.RegisterBackend("bolt2", &b)

		list, err := r.Endpoints()
		require.NoError(t, err)
		require.Len(t, list, 0)

		_, err = r.Create("bolt1", "a")
		require.NoError(t, err)

		list, err = r.Endpoints()
		require.NoError(t, err)
		require.Len(t, list, 1)

		_, err = r.Create("bolt2", "b")
		require.NoError(t, err)

		list, err = r.Endpoints()
		require.NoError(t, err)
		require.Len(t, list, 2)
	})
}
