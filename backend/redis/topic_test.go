package main

import (
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/require"
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
	list, err := redis.ByteSlices(topic.conn.Do("LRANGE", "topic:group", "0", "-1"))
	require.NoError(t, err)
	require.Len(t, list, 5)
	err = tp.Close()
	require.NoError(t, err)
}
