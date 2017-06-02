package main

import (
	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/backend/mongo/mongo"
)

const defaultURI = "mongodb://localhost:27017/lobby"

func main() {
	plugin := cli.NewPlugin("mongo")

	plugin.RunAsBackend(func() (lobby.Backend, error) {
		return mongo.NewBackend(defaultURI)
	})
}
