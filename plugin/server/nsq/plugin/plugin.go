package main

import (
	"fmt"
	"sync"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	lobbynsq "github.com/asdine/lobby/plugin/server/nsq"
	nsq "github.com/nsqio/go-nsq"
)

const (
	defaultNSQLookupdAddr = "127.0.0.1:4161"
	defaultTopic          = "lobby"
	defaultChannel        = "test"
)

var (
	ch chan struct{}
	wg sync.WaitGroup
)

// Config of the plugin
type Config struct {
	NSQLookupdAddr string
	Topic          string
	Channel        string
}

var cfg Config

func init() {
	wg.Add(1)
	ch = make(chan struct{})
}

// Name of the plugin
const Name = "nsq"

// Start plugin
func Start(reg lobby.Registry) error {
	defer wg.Done()
	if cfg.NSQLookupdAddr == "" {
		cfg.NSQLookupdAddr = defaultNSQLookupdAddr
	}

	if cfg.Topic == "" {
		cfg.Topic = defaultTopic
	}

	if cfg.Channel == "" {
		cfg.Channel = defaultChannel
	}

	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(cfg.Topic, cfg.Channel, config)
	if err != nil {
		return err
	}

	consumer.ChangeMaxInFlight(200)

	handler := lobbynsq.Handler{
		Registry: reg,
	}

	consumer.AddConcurrentHandlers(
		&handler,
		20,
	)

	fmt.Printf("Listening for NSQ messages\n")

	err = consumer.ConnectToNSQLookupds([]string{cfg.NSQLookupdAddr})
	if err != nil {
		return err
	}

	<-ch
	consumer.Stop()
	return nil
}

// Stop plugin
func Stop() error {
	close(ch)
	wg.Wait()
	return nil
}

func main() {
	cli.RunPlugin(Name, Start, Stop, &cfg)
}
