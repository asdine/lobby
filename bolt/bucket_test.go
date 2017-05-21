package bolt_test

import (
	"fmt"
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

func TestBucketPage(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := bolt.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)
	defer b.Close()

	for i := 0; i < 20; i++ {
		_, err := b.Put(fmt.Sprintf("%d", i), []byte("Value"))
		require.NoError(t, err)
	}

	list, err := b.Page(0, 0)
	require.NoError(t, err)
	require.Len(t, list, 0)

	list, err = b.Page(0, 10)
	require.NoError(t, err)
	require.Len(t, list, 0)

	list, err = b.Page(1, 5)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, "0", list[0].Key)
	require.Equal(t, "4", list[4].Key)

	list, err = b.Page(1, 25)
	require.NoError(t, err)
	require.Len(t, list, 20)
	require.Equal(t, "0", list[0].Key)
	require.Equal(t, "19", list[19].Key)

	list, err = b.Page(2, 5)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, "5", list[0].Key)
	require.Equal(t, "9", list[4].Key)

	list, err = b.Page(2, 15)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, "15", list[0].Key)
	require.Equal(t, "19", list[4].Key)

	list, err = b.Page(3, 15)
	require.NoError(t, err)
	require.Len(t, list, 0)

	// all
	list, err = b.Page(1, -1)
	require.NoError(t, err)
	require.Len(t, list, 20)
	require.Equal(t, "0", list[0].Key)
	require.Equal(t, "19", list[19].Key)
}
