package bolt_test

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/stretchr/testify/require"
)

func TestBucketPut(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := bolt.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("1a")
	require.NoError(t, err)

	i1, err := b.Put("2a", []byte("Value"))
	require.NoError(t, err)
	require.Equal(t, "2a", i1.Key)
	require.Equal(t, []byte("Value"), i1.Value)

	i2, err := b.Get("2a")
	require.NoError(t, err)
	require.Equal(t, *i1, *i2)

	j, err := b.Put("2a", []byte("New Value"))
	require.NoError(t, err)
	require.Equal(t, []byte("New Value"), j.Value)

	err = b.Close()
	require.NoError(t, err)
}

func TestBucketGet(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := bolt.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)

	i, err := b.Put("id", []byte("Value"))
	require.NoError(t, err)

	j, err := b.Get(i.Key)
	require.NoError(t, err)
	require.Equal(t, i.Value, j.Value)

	_, err = b.Get("some id")
	require.Equal(t, lobby.ErrKeyNotFound, err)

	err = b.Close()
	require.NoError(t, err)
}

func TestBucketDelete(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := bolt.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)

	i, err := b.Put("id", []byte("Value"))
	require.NoError(t, err)

	_, err = b.Get(i.Key)
	require.NoError(t, err)

	err = b.Delete(i.Key)
	require.NoError(t, err)

	err = b.Delete(i.Key)
	require.Error(t, err)
	require.Equal(t, lobby.ErrKeyNotFound, err)

	err = b.Close()
	require.NoError(t, err)
}
