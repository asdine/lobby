package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/redis"
)

const defaultAddr = ":6379"

// Name of the plugin
const Name = "redis"

// Config of the plugin
type Config struct {
	Addr string
}

var cfg Config

// Backend creates a Redis backend.
func Backend() (lobby.Backend, error) {
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}

	return redis.NewBackend(cfg.Addr)
}

func main() {
	cli.RunBackend(Name, Backend, &cfg)
}
