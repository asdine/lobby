package nsq

import (
	"testing"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/mock"
	lobbypb "github.com/asdine/lobby/rpc/proto"
	"github.com/gogo/protobuf/proto"
	nsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/require"
)

func TestHandleMessage(t *testing.T) {
	var reg mock.Registry

	reg.BucketFn = func(name string) (lobby.Bucket, error) {
		require.Equal(t, "bucket", name)
		return &mock.Bucket{
			PutFn: func(key string, value []byte) (*lobby.Item, error) {
				require.Equal(t, "key", key)
				require.Equal(t, `"value"`, string(value))
				return nil, nil
			},
		}, nil
	}

	h := Handler{
		Registry: &reg,
	}

	item := lobbypb.NewItem{
		Bucket: "bucket",
		Item: &lobbypb.Item{
			Key:   "key",
			Value: []byte(`"value"`),
		},
	}

	body, err := proto.Marshal(&item)
	require.NoError(t, err)

	err = h.HandleMessage(&nsq.Message{
		Body: body,
	})
	require.NoError(t, err)
}
