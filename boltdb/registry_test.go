package boltdb_test

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/boltdb"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	pathStore, cleanupStore := preparePath(t, "backend.db")
	defer cleanupStore()

	s, err := boltdb.NewBackend(pathStore)
	require.NoError(t, err)
	defer s.Close()

	t.Run("create", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := boltdb.NewRegistry(pathReg)
		require.NoError(t, err)
		defer r.Close()

		err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrBackendNotFound, err)

		r.RegisterBackend("bolt1", s)
		r.RegisterBackend("bolt2", s)

		err = r.Create("bolt1", "a")
		require.NoError(t, err)

		err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrBucketAlreadyExists, err)

		err = r.Create("bolt1", "b")
		require.NoError(t, err)

		err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrBucketAlreadyExists, err)
	})

	t.Run("bucket", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := boltdb.NewRegistry(pathReg)
		require.NoError(t, err)
		defer r.Close()

		r.RegisterBackend("bolt1", s)
		r.RegisterBackend("bolt2", s)

		b, err := r.Bucket("")
		require.Equal(t, lobby.ErrBucketNotFound, err)

		b, err = r.Bucket("a")
		require.Equal(t, lobby.ErrBucketNotFound, err)

		err = r.Create("bolt1", "a")
		require.NoError(t, err)

		b, err = r.Bucket("a")
		require.NoError(t, err)
		require.NotNil(t, b)

		err = r.Create("bolt2", "b")
		require.NoError(t, err)

		b, err = r.Bucket("b")
		require.NoError(t, err)
		require.NotNil(t, b)

		err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrBucketAlreadyExists, err)
	})
}
