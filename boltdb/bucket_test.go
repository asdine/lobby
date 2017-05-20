package boltdb_test

import (
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/boltdb"
	"github.com/stretchr/testify/require"
)

func TestBucketSave(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := boltdb.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("1a")
	require.NoError(t, err)

	i1, err := b.Save("2a", []byte("Data"))
	require.NoError(t, err)
	require.Equal(t, "2a", i1.Key)
	require.Equal(t, []byte("Data"), i1.Data)

	i2, err := b.Get("2a")
	require.NoError(t, err)
	require.Equal(t, *i1, *i2)

	j, err := b.Save("2a", []byte("New Data"))
	require.NoError(t, err)
	require.Equal(t, []byte("New Data"), j.Data)

	err = b.Close()
	require.NoError(t, err)
}

func TestBucketGet(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := boltdb.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)

	i, err := b.Save("id", []byte("Data"))
	require.NoError(t, err)

	j, err := b.Get(i.Key)
	require.NoError(t, err)
	require.Equal(t, i.Data, j.Data)

	_, err = b.Get("some id")
	require.Equal(t, lobby.ErrKeyNotFound, err)

	err = b.Close()
	require.NoError(t, err)
}

func TestBucketDelete(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	s, err := boltdb.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)

	i, err := b.Save("id", []byte("Data"))
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

	s, err := boltdb.NewBackend(path)
	require.NoError(t, err)

	b, err := s.Bucket("a")
	require.NoError(t, err)
	defer b.Close()

	for i := 0; i < 20; i++ {
		_, err := b.Save(fmt.Sprintf("%d", i), []byte("Data"))
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
