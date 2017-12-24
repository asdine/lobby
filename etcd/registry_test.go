package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/asdine/lobby/etcd/internal"
	"github.com/asdine/lobby/mock"
	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379"}
)

func etcdHelper(t *testing.T) (*clientv3.Client, func()) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	return cli, func() {
		_, err := cli.Delete(context.Background(), "", clientv3.WithPrefix())
		assert.NoError(t, err)
		cli.Close()
	}
}

func TestEtcdRegistry(t *testing.T) {
	client, cleanup := etcdHelper(t)
	defer cleanup()

	createTopics(t, client, "lobby-tests", 5)

	reg, err := NewRegistry(client, "lobby-tests")
	require.NoError(t, err)
	require.Len(t, reg.topics, 5)

	reg.RegisterBackend("backend", new(mock.Backend))
	err = reg.Create("backend", "sometopic")
	require.NoError(t, err)
	require.Len(t, reg.topics, 6)

	_, err = reg.Topic("sometopic")
	require.NoError(t, err)

	err = reg.Close()
	require.NoError(t, err)
}

func createTopics(t *testing.T, client *clientv3.Client, namespace string, count int) {
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s/topics/topic-%d", namespace, i)
		raw, err := proto.Marshal(&internal.Topic{
			Name:    fmt.Sprintf("topic-%d", i),
			Backend: "backend",
		})
		require.NoError(t, err)
		_, err = client.Put(context.Background(), key, string(raw))
		require.NoError(t, err)
	}
}
