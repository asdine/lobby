package main

import (
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	nsq "github.com/nsqio/go-nsq"
)

const (
	defaultNSQAddr = "127.0.0.1:4150"
)

// Config of the plugin.
type Config struct {
	NSQAddr string
}

func main() {
	var cfg Config

	cli.RunBackend("nsq", func() (lobby.Backend, error) {
		if cfg.NSQAddr == "" {
			cfg.NSQAddr = defaultNSQAddr
		}

		return NewBackend(cfg.NSQAddr)
	}, &cfg)
}

var _ lobby.Backend = new(Backend)

// NewBackend returns a NSQ backend.
func NewBackend(addr string) (*Backend, error) {
	var err error

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil, err
	}

	p.SetLogger(log.New(os.Stderr, "", 0), nsq.LogLevelInfo)

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
