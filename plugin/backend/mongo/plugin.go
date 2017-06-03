package main

import (
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/mongo/mongo"
)

const defaultURI = "mongodb://localhost:27017/lobby"

func main() {
	plugin := cli.NewPlugin("mongo")

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = defaultURI
	}

	plugin.RunAsBackend(func() (lobby.Backend, error) {
		return mongo.NewBackend(uri)
	})
}
