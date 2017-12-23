package nsq

import (
	"errors"

	"github.com/asdine/lobby"
	lobbypb "github.com/asdine/lobby/rpc/proto"
	"github.com/gogo/protobuf/proto"
	nsq "github.com/nsqio/go-nsq"
)

// Handler is a NSQ handler.
type Handler struct {
	Registry lobby.Registry
}

// HandleMessage decodes proto messages and sends them to the right topic.
func (h *Handler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return errors.New("body is blank re-enqueue message")
	}

	var newMessage lobbypb.NewMessage
	err := proto.Unmarshal(m.Body, &newMessage)
	if err != nil {
		return err
	}

	topic, err := h.Registry.Topic(newMessage.Topic)
	if err != nil {
		return err
	}

	return topic.Send(&lobby.Message{
		Group: newMessage.Message.Group,
		Value: newMessage.Message.Value,
	})
}
