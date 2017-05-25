package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/asdine/lobby"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func newApp() *app {
	a := app{
		in:  os.Stdin,
		out: os.Stdout,
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	defaultConfigDir := path.Join(home, ".config/lobby")

	c := cli.NewApp()
	c.Name = "lobby"
	c.Version = lobby.Version
	c.Before = a.init
	c.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config-dir",
			Usage:       "Path to a directory to read and store Lobby configuration",
			Destination: &a.ConfigDir,
			Value:       defaultConfigDir,
		},
		cli.StringFlag{
			Name:        "data-dir",
			Usage:       "Path to a directory to read and store Lobby data",
			Destination: &a.DataDir,
			Value:       path.Join(defaultConfigDir, "data"),
		},
		cli.StringFlag{
			Name:        "socket-dir",
			Usage:       "Path to a directory to read and store Lobby sockets",
			Destination: &a.SocketDir,
			Value:       path.Join(defaultConfigDir, "sockets"),
		},
	}

	a.App = c
	return &a
}

type app struct {
	*cli.App

	in        io.Reader
	out       io.Writer
	registry  lobby.Registry
	ConfigDir string
	DataDir   string
	SocketDir string
}

func (a *app) init(c *cli.Context) error {
	paths := []string{
		a.ConfigDir,
		a.DataDir,
		a.SocketDir,
	}

	for _, path := range paths {
		err := initDir(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func initDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "Can't create directory %s", path)
		}

		return nil
	}

	if !fi.Mode().IsDir() {
		return fmt.Errorf("'%s' is not a valid directory", path)
	}

	return nil
}
