package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
)

const defaultAddr = ":6379"

// Config of the plugin
type Config struct {
	Addr string
}

func main() {
	var cfg Config

	cli.RunBackend("redis", func() (lobby.Backend, error) {
		if cfg.Addr == "" {
			cfg.Addr = defaultAddr
		}

		return NewBackend(cfg.Addr)
	}, &cfg)
}
