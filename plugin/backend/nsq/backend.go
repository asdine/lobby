package nsq

import (
	"github.com/asdine/lobby"
	nsq "github.com/nsqio/go-nsq"
)

var _ lobby.Backend = new(Backend)

// NewBackend returns a NSQ backend.
func NewBackend(addr string) (*Backend, error) {
	var err error

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil, err
	}

	return &Backend{
		producer: p,
	}, nil
}

// Backend is a NSQ backend.
type Backend struct {
	producer *nsq.Producer
}

// Topic returns the topic associated with the given name.
func (s *Backend) Topic(name string) (lobby.Topic, error) {
	return lobby.TopicFunc(func(m *lobby.Message) error {
		return s.producer.Publish(name, m.Value)
	}), nil
}

// Close NSQ connection.
func (s *Backend) Close() error {
	s.producer.Stop()
	return nil
}
