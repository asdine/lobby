package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/mongo"
)

const defaultURI = "mongodb://localhost:27017/lobby"

// Name of the plugin
const Name = "mongo"

// Config of the plugin
type Config struct {
	URI string `toml:"uri"`
}

var cfg Config

// Backend creates a MongoDB backend.
func Backend() (lobby.Backend, error) {
	if cfg.URI == "" {
		cfg.URI = defaultURI
	}

	return mongo.NewBackend(cfg.URI)
}

func main() {
	cli.RunBackend(Name, Backend, &cfg)
}
