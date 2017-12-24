package bolt

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt/boltpb"
	"github.com/asdine/storm"
	"github.com/pkg/errors"
)

var _ lobby.Topic = new(Topic)

// NewTopic returns a Topic
func NewTopic(node storm.Node) *Topic {
	return &Topic{
		node: node,
	}
}

// Topic is a BoltDB implementation of a topic.
type Topic struct {
	node storm.Node
}

// Send a message to the topic.
func (t *Topic) Send(message *lobby.Message) error {
	tx, err := t.node.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create bolt transaction")
	}
	defer tx.Rollback()

	err = tx.Save(&boltpb.Message{
		Group: message.Group,
		Value: message.Value,
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit bolt transaction")
	}

	return nil
}

// Close the topic session.
func (t *Topic) Close() error {
	return nil
}
