package main

import (
	"fmt"
	"os"

	"github.com/asdine/lobby"
	lobbynsq "github.com/asdine/lobby/builtin/server/nsq"
	"github.com/asdine/lobby/cli"
	nsq "github.com/nsqio/go-nsq"
)

const (
	defaultNSQLookupAddr = "127.0.0.1:4161"
	defaultTopic         = "lobby"
	defaultChannel       = "test"
)

var (
	addr    string
	topic   string
	channel string
	ch      chan struct{}
)

func init() {
	addr = os.Getenv("NSQLOOKUPD_ADDR")
	if addr == "" {
		addr = defaultNSQLookupAddr
	}

	topic = os.Getenv("NSQ_TOPIC")
	if topic == "" {
		topic = defaultTopic
	}

	channel = os.Getenv("NSQ_CHANNEL")
	if channel == "" {
		channel = defaultChannel
	}

	ch = make(chan struct{})
}

// Name of the plugin
const Name = "nsq"

// Start plugin
func Start(reg lobby.Registry) error {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
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

	err = consumer.ConnectToNSQLookupds([]string{addr})
	if err != nil {
		return err
	}

	<-ch
	consumer.Stop()
	return nil
}

// Stop plugin
func Stop() error {
	ch <- struct{}{}
	return nil
}

func main() {
	cli.RunPlugin(Name, Start, Stop)
}
