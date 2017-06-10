package main

import (
	"log"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/builtin/backend/mongo"
	"github.com/asdine/lobby/cli"
)

const defaultURI = "mongodb://localhost:27017/lobby"

// Name of the plugin
const Name = "mongo"

// Backend creates a MongoDB backend.
func Backend() (lobby.Backend, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = defaultURI
	}

	return mongo.NewBackend(uri)
}

func main() {
	backend, err := Backend()
	if err != nil {
		log.Fatal(err)
	}

	cli.RunBackend(Name, backend)
}
