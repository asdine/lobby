package bolt_test

import (
	"io/ioutil"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/log"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	pathStore, cleanupStore := preparePath(t, "backend.db")
	defer cleanupStore()

	s, err := bolt.NewBackend(pathStore)
	require.NoError(t, err)
	defer s.Close()

	t.Run("create", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := bolt.NewRegistry(pathReg, log.New(log.Output(ioutil.Discard)))
		require.NoError(t, err)
		defer r.Close()

		err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrBackendNotFound, err)

		r.RegisterBackend("bolt1", s)
		r.RegisterBackend("bolt2", s)

		err = r.Create("bolt1", "a")
		require.NoError(t, err)

		err = r.Create("bolt1", "a")
		require.Equal(t, lobby.ErrTopicAlreadyExists, err)

		err = r.Create("bolt1", "b")
		require.NoError(t, err)

		err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrTopicAlreadyExists, err)
	})

	t.Run("topic", func(t *testing.T) {
		pathReg, cleanupReg := preparePath(t, "reg.db")
		defer cleanupReg()
		r, err := bolt.NewRegistry(pathReg, log.New(log.Output(ioutil.Discard)))
		require.NoError(t, err)
		defer r.Close()

		r.RegisterBackend("bolt1", s)
		r.RegisterBackend("bolt2", s)

		b, err := r.Topic("")
		require.Equal(t, lobby.ErrTopicNotFound, err)

		b, err = r.Topic("a")
		require.Equal(t, lobby.ErrTopicNotFound, err)

		err = r.Create("bolt1", "a")
		require.NoError(t, err)

		b, err = r.Topic("a")
		require.NoError(t, err)
		require.NotNil(t, b)

		err = r.Create("bolt2", "b")
		require.NoError(t, err)

		b, err = r.Topic("b")
		require.NoError(t, err)
		require.NotNil(t, b)

		err = r.Create("bolt2", "a")
		require.Equal(t, lobby.ErrTopicAlreadyExists, err)
	})
}
