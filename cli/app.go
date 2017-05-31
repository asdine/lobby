package cli

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/plugin"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v1"
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

	defaultPluginDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	c := cli.NewApp()
	c.Name = "lobby"
	c.Version = lobby.Version
	c.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config-dir",
			Usage:       "Path to a directory to read and store Lobby configuration and data",
			Destination: &a.ConfigDir,
			Value:       defaultConfigDir,
		},
		cli.StringFlag{
			Name:        "plugin-dir",
			Usage:       "Path to a directory to read Lobby plugins",
			Destination: &a.PluginDir,
			Value:       defaultPluginDir,
		},
		cli.StringSliceFlag{
			Name:  "backend",
			Usage: "Name of the backend to use",
		},
		cli.StringSliceFlag{
			Name:  "server",
			Usage: "Name of the server to run",
		},
	}

	c.Before = func(c *cli.Context) error {
		a.backendList = c.StringSlice("backend")
		a.serverList = c.StringSlice("server")
		a.DataDir = path.Join(defaultConfigDir, "data")
		a.SocketDir = path.Join(defaultConfigDir, "sockets")
		return a.init()
	}

	c.After = func(c *cli.Context) error {
		return a.closePlugins()
	}

	a.App = c
	return &a
}

type app struct {
	*cli.App

	in          io.Reader
	out         io.Writer
	registry    lobby.Registry
	ConfigDir   string
	DataDir     string
	SocketDir   string
	PluginDir   string
	Backends    []plugin.Backend
	Servers     []plugin.Plugin
	backendList []string
	serverList  []string
}

func (a *app) init() error {
	err := a.initDirectories()
	if err != nil {
		return err
	}

	return a.loadPlugins()
}

func (a *app) initDirectories() error {
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

func (a *app) loadPlugins() error {
	var err error
	a.Backends = make([]plugin.Backend, len(a.backendList))

	for i, name := range a.backendList {
		a.Backends[i], err = plugin.LoadBackend(name, path.Join(a.PluginDir, fmt.Sprintf("lobby-%s", name)), a.ConfigDir)
		if err != nil {
			return err
		}
	}

	a.Servers = make([]plugin.Plugin, len(a.serverList))

	for i, name := range a.serverList {
		a.Servers[i], err = plugin.LoadServer(name, path.Join(a.PluginDir, fmt.Sprintf("lobby-%s", name)), a.ConfigDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) closePlugins() error {
	for _, p := range a.Servers {
		err := p.Close()
		if err != nil {
			return err
		}
	}

	for _, p := range a.Backends {
		err := p.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) runServers(servers map[net.Listener]lobby.Server) error {
	var wg sync.WaitGroup

	for l, srv := range servers {
		wg.Add(1)
		go func(l net.Listener, srv lobby.Server) {
			defer wg.Done()
			fmt.Fprintf(a.out, "Listening %s requests on %s.\n", srv.Name(), l.Addr().String())
			srv.Serve(l)
		}(l, srv)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	var lastErr error
	for _, srv := range servers {
		if err := srv.Stop(); err != nil {
			lastErr = err
		}
	}

	wg.Wait()
	return lastErr
}
