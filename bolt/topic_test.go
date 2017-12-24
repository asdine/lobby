package bolt_test

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/bolt/boltpb"
	"github.com/stretchr/testify/require"
)

func TestTopicSend(t *testing.T) {
	path, cleanup := preparePath(t, "store.db")
	defer cleanup()

	bk, err := bolt.NewBackend(path)
	require.NoError(t, err)

	tp, err := bk.Topic("1a")
	require.NoError(t, err)

	err = tp.Send(&lobby.Message{
		Group: "2a",
		Value: []byte("Value"),
	})
	require.NoError(t, err)

	var m []boltpb.Message
	err = bk.DB.From("1a").Find("Group", "2a", &m)
	require.NoError(t, err)
	require.Len(t, m, 1)

	err = tp.Send(&lobby.Message{
		Group: "2a",
		Value: []byte("New Value"),
	})
	require.NoError(t, err)

	err = bk.DB.From("1a").Find("Group", "2a", &m)
	require.NoError(t, err)
	require.Len(t, m, 2)

	err = tp.Close()
	require.NoError(t, err)
}
