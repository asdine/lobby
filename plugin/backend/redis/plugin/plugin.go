package main

import (
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/redis"
)

const defaultAddr = ":6379"

// Name of the plugin
const Name = "redis"

// Backend creates a Redis backend.
func Backend() (lobby.Backend, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	return redis.NewBackend(addr)
}

func main() {
	backend, err := Backend()
	if err != nil {
		log.Fatal(err)
	}

	cli.RunBackend(Name, backend)
}
