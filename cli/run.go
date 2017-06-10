package cli

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/rpc"
	"github.com/spf13/cobra"
)

const (
	defaultAddr = ":5656"
)

func newRunCmd(a *app) *cobra.Command {
	r := runCmd{app: a}

	cmd := cobra.Command{
		Use:  "run",
		RunE: r.run,
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return r.closePlugins()
		},
	}

	cmd.Flags().StringSliceVar(&r.backendList, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringSliceVar(&r.serverList, "server", nil, "Name of the server to run")

	return &cmd
}

type runCmd struct {
	app         *app
	mainSrv     lobby.Server
	plugins     []lobby.Plugin
	backends    map[string]lobby.Backend
	backendList []string
	serverList  []string
}

func (r *runCmd) run(cmd *cobra.Command, args []string) error {
	return r.runMainServer()
}

func (r *runCmd) loadBackendPlugins() error {
	r.backends = make(map[string]lobby.Backend)

	for _, name := range r.backendList {
		bck, plg, err := rpc.LoadBackendPlugin(name, path.Join(r.app.PluginDir, fmt.Sprintf("lobby-%s", name)), r.app.ConfigDir)
		if err != nil {
			return err
		}

		r.backends[name] = bck
		r.plugins = append(r.plugins, plg)
	}

	return nil
}

func (r *runCmd) loadServerPlugins() error {
	for _, name := range r.serverList {
		plg, err := rpc.LoadPlugin(name, path.Join(r.app.PluginDir, fmt.Sprintf("lobby-%s", name)), r.app.ConfigDir)
		if err != nil {
			return err
		}

		r.plugins = append(r.plugins, plg)
	}

	return nil
}

func (r *runCmd) closePlugins() error {
	for _, p := range r.plugins {
		err := p.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *runCmd) runMainServer() error {
	err := r.loadBackendPlugins()
	if err != nil {
		return err
	}

	dataPath := path.Join(r.app.DataDir, "bolt")
	registryPath := path.Join(dataPath, "registry.db")
	backendPath := path.Join(dataPath, "backend.db")

	err = initDir(dataPath)
	if err != nil {
		return err
	}

	// Creating default registry.
	reg, err := bolt.NewRegistry(registryPath)
	if err != nil {
		return err
	}

	// Creating default backend.
	bck, err := bolt.NewBackend(backendPath)
	if err != nil {
		return err
	}
	reg.RegisterBackend("bolt", bck)

	// Loading backends from plugins.
	for name, bck := range r.backends {
		reg.RegisterBackend(name, bck)
	}

	srv := rpc.NewServer(rpc.WithBucketService(reg), rpc.WithRegistryService(reg))

	// listening on specific port
	l, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		return err
	}

	// listening on unix socket
	lsock, err := net.Listen("unix", path.Join(r.app.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}

	err = r.loadServerPlugins()
	if err != nil {
		return err
	}

	return runServers(r.app.out, map[net.Listener]lobby.Server{
		l:     srv,
		lsock: srv,
	}, func() error {
		err := bck.Close()
		if err != nil {
			return err
		}

		return reg.Close()
	})
}

func runServers(out io.Writer, servers map[net.Listener]lobby.Server, beforeStop ...func() error) error {
	var wg sync.WaitGroup

	for l, srv := range servers {
		wg.Add(1)
		go func(l net.Listener, srv lobby.Server) {
			defer wg.Done()
			fmt.Fprintf(out, "Listening %s requests on %s.\n", srv.Name(), l.Addr().String())
			srv.Serve(l)
		}(l, srv)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Fprintf(out, "\nStopping servers...")
	for _, fn := range beforeStop {
		err := fn()
		if err != nil {
			return err
		}
	}

	var lastErr error
	for _, srv := range servers {
		if err := srv.Stop(); err != nil {
			lastErr = err
		}
	}

	wg.Wait()
	if lastErr == nil {
		fmt.Fprintf(out, " OK\n")
	}

	return lastErr
}
