package main

import (
	"path"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/cli"
)

func main() {
	plugin := cli.NewPlugin("bolt")

	plugin.RunAsBackend(func() (lobby.Backend, error) {
		return bolt.NewBackend(path.Join(plugin.DataDir, "bolt", "backend.db"))
	})
}
