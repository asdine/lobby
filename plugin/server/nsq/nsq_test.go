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

	reg.TopicFn = func(name string) (lobby.Topic, error) {
		require.Equal(t, "topic", name)
		return &mock.Topic{
			SendFn: func(m *lobby.Message) error {
				require.Equal(t, "group", m.Group)
				require.Equal(t, `"value"`, string(m.Value))
				return nil
			},
		}, nil
	}

	h := Handler{
		Registry: &reg,
	}

	item := lobbypb.NewMessage{
		Topic: "topic",
		Message: &lobbypb.Message{
			Group: "group",
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
