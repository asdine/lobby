package mongo

import (
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
)

func TestTopicSend(t *testing.T) {
	backend, cleanup := getBackend(t)
	defer cleanup()

	tp, err := backend.Topic("topic")
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		err = tp.Send(&lobby.Message{
			Group: "group",
			Value: []byte(fmt.Sprintf("Value%d", i)),
		})
		require.NoError(t, err)
	}

	topic := tp.(*Topic)
	col := topic.session.DB("").C(colMessages)
	var list []message
	err = col.Find(bson.M{"group": "group"}).All(&list)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, []byte("Value0"), list[0].Value)
	err = tp.Close()
	require.NoError(t, err)
}
