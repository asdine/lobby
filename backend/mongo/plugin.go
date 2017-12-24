package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
)

const defaultURI = "mongodb://localhost:27017/lobby"

// Config of the plugin
type Config struct {
	URI string `toml:"uri"`
}

func main() {
	var cfg Config

	cli.RunBackend("mongo", func() (lobby.Backend, error) {
		if cfg.URI == "" {
			cfg.URI = defaultURI
		}

		return NewBackend(cfg.URI)
	}, &cfg)
}
