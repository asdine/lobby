package main

import (
	"net"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/http"
)

const defaultAddr = ":5657"

func main() {
	plugin := cli.NewPlugin("http")

	plugin.RunAsServer(func(reg lobby.Registry) (net.Listener, lobby.Server, error) {
		l, err := net.Listen("tcp", defaultAddr)
		if err != nil {
			return nil, nil, err
		}

		srv := http.NewServer(http.NewHandler(reg))

		return l, srv, nil
	})
}
