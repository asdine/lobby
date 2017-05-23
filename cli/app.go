package cli

import (
	"errors"
	"io"
	"log"
	"os"
	"path"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/asdine/lobby"
	homedir "github.com/mitchellh/go-homedir"
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

	c := cli.NewApp()
	c.Name = "lobby"
	c.Version = lobby.Version
	c.Before = a.init
	c.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config-dir",
			Usage:       "Path to a directory to read and store Lobby configuration",
			Destination: &a.configDir,
			Value:       path.Join(home, ".config/lobby"),
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
	configDir string
}

func (a *app) init(c *cli.Context) error {
	err := a.initConfigDir()
	if err != nil {
		return err
	}

	return nil
}

func (a *app) initConfigDir() error {
	fi, err := os.Stat(a.configDir)
	if err != nil {
		return os.Mkdir(a.configDir, 0755)
	}

	if !fi.Mode().IsDir() {
		return errors.New("Config directory must be a valid directory")
	}

	return nil
}
