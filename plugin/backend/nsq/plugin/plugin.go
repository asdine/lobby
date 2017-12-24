package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/nsq"
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

		return nsq.NewBackend(cfg.NSQAddr)
	}, &cfg)
}
