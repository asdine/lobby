package rpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	"github.com/asdine/lobby/rpc/proto"
	"github.com/stretchr/testify/require"
)

func TestPut(t *testing.T) {
	var r mock.Registry

	var i int
	r.BucketFn = func(name string) (lobby.Bucket, error) {
		require.Equal(t, "bucket", name)

		return &mock.Bucket{
			SaveFn: func(key string, data []byte) (*lobby.Item, error) {
				require.Equal(t, fmt.Sprintf("key%d", i+1), key)
				require.Equal(t, fmt.Sprintf(`"value%d"`, i+1), string(data))
				i++
				return &lobby.Item{
					Key:  key,
					Data: data,
				}, nil
			},
		}, nil
	}

	conn, cleanup := newServer(t, &r)
	defer cleanup()

	client := proto.NewBucketServiceClient(conn)

	stream, err := client.Put(context.Background())
	require.NoError(t, err)

	for j := 0; j < 5; j++ {
		err = stream.Send(&proto.NewItem{
			Item: &proto.Item{
				Key:   fmt.Sprintf("key%d", j+1),
				Value: []byte(fmt.Sprintf("value%d", j+1)),
			},
			Bucket: "bucket",
		})
		require.NoError(t, err)
	}

	summary, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.Equal(t, int32(5), summary.ItemCount)
}
