package main

import (
	"errors"
	"net"
	"os"
	"path"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/rpc"
)

func main() {
	var dataDir string
	var addr string

	app := cli.NewApp()
	app.Name = "bolt"
	app.Usage = "Boltdb backend"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "data-dir",
			Destination: &dataDir,
		},
		cli.StringFlag{
			Name:        "addr",
			Destination: &addr,
		},
	}

	app.Action = func(c *cli.Context) error {
		if addr == "" || dataDir == "" {
			return errors.New("invalid arguments")
		}

		bck, err := bolt.NewBackend(path.Join(dataDir, "backend.db"))
		if err != nil {
			return err
		}
		defer bck.Close()

		l, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		server := rpc.NewServer(rpc.WithBucketService(bck))
		return server.Serve(l)
	}

	app.Run(os.Args)
}
