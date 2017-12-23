package mongo

import (
	"encoding/json"

	"github.com/asdine/lobby"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
)

const colMessages = "messages"

type message struct {
	ID    string      `bson:"_id,omitempty"`
	Topic string      `bson:"topic"`
	Group string      `bson:"group"`
	Value interface{} `bson:"value"`
}

var _ lobby.Topic = new(Topic)

// NewTopic returns a MongoDB Topic.
func NewTopic(session *mgo.Session, name string) *Topic {
	return &Topic{
		session: session,
		name:    name,
	}
}

// Topic is a MongoDB implementation of a topic.
type Topic struct {
	session *mgo.Session
	name    string
}

// Send a message to the topic.
func (t *Topic) Send(m *lobby.Message) error {
	col := t.session.DB("").C(colMessages)

	var raw interface{}

	valid, err := ValidateBytes(m.Value)
	if err == nil {
		err := json.Unmarshal(valid, &raw)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal json")
		}
	} else {
		raw = m.Value
	}

	err = col.Insert(&message{Group: m.Group, Topic: t.name, Value: raw})
	if err != nil {
		return errors.Wrap(err, "failed to insert of update")
	}

	return nil
}

// Close the topic session.
func (t *Topic) Close() error {
	t.session.Close()
	return nil
}

// ValidateBytes checks if the data is valid json.
func ValidateBytes(data []byte) ([]byte, error) {
	var i json.RawMessage

	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	return i, nil
}
