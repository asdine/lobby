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

// HandleMessage decodes proto messages and sends them to the right bucket.
func (h *Handler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return errors.New("body is blank re-enqueue message")
	}

	var newItem lobbypb.NewItem
	err := proto.Unmarshal(m.Body, &newItem)
	if err != nil {
		return err
	}

	bucket, err := h.Registry.Bucket(newItem.Bucket)
	if err != nil {
		return err
	}

	_, err = bucket.Put(newItem.Item.Key, newItem.Item.Value)
	return err
}
