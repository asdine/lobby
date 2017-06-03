package main

import (
	"net"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/server/http/http"
)

const defaultAddr = ":5657"

func main() {
	plugin := cli.NewPlugin("http")

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	plugin.RunAsServer(func(reg lobby.Registry) (net.Listener, lobby.Server, error) {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, nil, err
		}

		srv := http.NewServer(http.NewHandler(reg))

		return l, srv, nil
	})
}
