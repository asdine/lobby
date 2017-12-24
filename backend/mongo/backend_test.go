package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getBackend(t *testing.T) (*Backend, func()) {
	bck, err := NewBackend("mongodb://localhost:27017/test-db")
	require.NoError(t, err)

	return bck, func() {
		_, err := bck.session.DB("").C(colMessages).RemoveAll(nil)
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
	require.NotNil(t, topic.(*Topic).session)

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
