package main

import (
	"fmt"
	"net"
	"os"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli"
	"github.com/asdine/lobby/plugin/server/http"
)

const defaultAddr = ":5657"

var (
	srv  lobby.Server
	addr string
)

func init() {
	addr = os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = defaultAddr
	}
}

// Name of the plugin
const Name = "http"

// Start plugin
func Start(reg lobby.Registry) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	fmt.Printf("Listening http requests on %s\n", l.Addr())
	srv = http.NewServer(http.NewHandler(reg))

	return srv.Serve(l)
}

// Stop plugin
func Stop() error {
	return srv.Stop()
}

func main() {
	cli.RunPlugin(Name, Start, Stop)
}
