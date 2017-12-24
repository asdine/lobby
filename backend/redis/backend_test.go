package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getBackend(t *testing.T) (*Backend, func()) {
	bck, err := NewBackend(":6379")
	require.NoError(t, err)

	return bck, func() {
		conn := bck.pool.Get()
		defer conn.Close()
		_, err := conn.Do("FLUSHDB")
		if err != nil {
			t.Error(err)
		}
	}
}

func TestBackend(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	topic, err := backend.Topic("a")
	require.NoError(t, err)
	require.NotNil(t, topic)
	require.NotNil(t, topic.(*Topic).conn)

	err = topic.Close()
	require.NoError(t, err)

	b1, err := backend.Topic("a")
	require.NoError(t, err)

	b2, err := backend.Topic("b")
	require.NoError(t, err)

	err = b1.Close()
	require.NoError(t, err)

	err = b2.Close()
	require.NoError(t, err)
}
