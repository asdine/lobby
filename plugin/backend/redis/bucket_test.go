package redis

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/stretchr/testify/require"
)

func TestBucketPut(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	b, err := backend.Bucket("1a")
	require.NoError(t, err)

	i1, err := b.Put("2a", []byte(`"Value"`))
	require.NoError(t, err)
	require.Equal(t, "2a", i1.Key)
	require.Equal(t, []byte(`"Value"`), i1.Value)

	i2, err := b.Get("2a")
	require.NoError(t, err)
	require.Equal(t, *i1, *i2)

	j, err := b.Put("2a", []byte(`"New Value"`))
	require.NoError(t, err)
	require.Equal(t, []byte(`"New Value"`), j.Value)

	err = b.Close()
	require.NoError(t, err)
}

func TestBucketGet(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	b1, err := backend.Bucket("b1")
	require.NoError(t, err)
	b2, err := backend.Bucket("b2")
	require.NoError(t, err)

	i1, err := b1.Put("id", []byte(`"Value1"`))
	require.NoError(t, err)
	i2, err := b2.Put("id", []byte(`"Value2"`))
	require.NoError(t, err)

	j1, err := b1.Get(i1.Key)
	require.NoError(t, err)
	require.Equal(t, i1.Value, j1.Value)

	j2, err := b2.Get(i2.Key)
	require.NoError(t, err)
	require.Equal(t, i2.Value, j2.Value)

	i2, err = b2.Put("some id", []byte(`"Value2"`))
	require.NoError(t, err)

	_, err = b1.Get("some id")
	require.Equal(t, lobby.ErrKeyNotFound, err)

	err = b1.Close()
	require.NoError(t, err)

	err = b2.Close()
	require.NoError(t, err)
}

func TestBucketDelete(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	b, err := backend.Bucket("a")
	require.NoError(t, err)

	i, err := b.Put("id", []byte(`"Value"`))
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
