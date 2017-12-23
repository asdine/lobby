package redis

import (
	"github.com/asdine/lobby"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

var _ lobby.Topic = new(Topic)

// NewTopic returns a Redis Topic.
func NewTopic(conn redis.Conn, name string) *Topic {
	return &Topic{
		conn: conn,
		name: name,
	}
}

// Topic is a Redis implementation of a topic.
type Topic struct {
	conn redis.Conn
	name string
}

// Send message to the topic.
func (t *Topic) Send(m *lobby.Message) error {
	name := t.name
	if m.Group != "" {
		name += ":" + m.Group
	}

	_, err := t.conn.Do("RPUSH", name, m.Value)
	return errors.Wrapf(err, "failed to send message '%s'", name)
}

// Close the topic connection.
func (t *Topic) Close() error {
	return t.conn.Close()
}
