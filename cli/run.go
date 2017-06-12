package cli

import (
	"fmt"
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

func newRunCmd(a *app) *cobra.Command {
	r := runCmd{app: a}
	a.out = lobby.NewPrefixWriter(fmt.Sprintf("[lobby]\t"), a.out)

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the lobby server",
		RunE:  r.run,
	}

	cmd.Flags().StringSliceVar(&r.backendList, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringSliceVar(&r.serverList, "server", nil, "Name of the server to run")

	return &cmd
}

type runCmd struct {
	app         *app
	mainSrv     lobby.Server
	plugins     []lobby.Plugin
	backendList []string
	serverList  []string
}

func (r *runCmd) run(cmd *cobra.Command, args []string) error {
	return r.runMainServer()
}

func (r *runCmd) runMainServer() error {
	dataPath := path.Join(r.app.DataDir, "bolt")
	registryPath := path.Join(dataPath, "registry.db")
	backendPath := path.Join(dataPath, "backend.db")

	err := initDir(dataPath)
	if err != nil {
		return err
	}

	// Creating default backend.
	bck, err := bolt.NewBackend(backendPath)
	if err != nil {
		return err
	}

	// Creating default registry.
	reg, err := bolt.NewRegistry(registryPath)
	if err != nil {
		return err
	}
	defer reg.Close()

	reg.RegisterBackend("bolt", bck)

	wg, srv, err := r.runServer(reg)
	if err != nil {
		return err
	}

	err = r.loadPlugins(reg)
	if err != nil {
		r.closeAll(srv)
		wg.Wait()
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println()

	r.closeAll(srv)
	wg.Wait()
	return nil
}

func (r *runCmd) closeAll(srv lobby.Server) {
	fmt.Fprintf(r.app.out, "Stopping plugins...")
	if err := r.closePlugins(); err != nil {
		fmt.Fprintf(r.app.out, " Error: %s\n", err.Error())
	} else {
		fmt.Fprintf(r.app.out, " OK\n")
	}

	fmt.Fprintf(r.app.out, "Stopping lobby...")
	if err := srv.Stop(); err != nil {
		fmt.Fprintf(r.app.out, " Error: %s\n", err.Error())
	} else {
		fmt.Fprintf(r.app.out, " OK\n")
	}
}

func (r *runCmd) loadPlugins(reg lobby.Registry) error {
	err := r.loadBackendPlugins(reg)
	if err != nil {
		return err
	}

	return r.loadServerPlugins()
}

func (r *runCmd) loadBackendPlugins(reg lobby.Registry) error {
	for _, name := range r.backendList {
		bck, plg, err := rpc.LoadBackendPlugin(name, path.Join(r.app.PluginDir, fmt.Sprintf("lobby-%s", name)), r.app.ConfigDir)
		if err != nil {
			return err
		}

		reg.RegisterBackend(name, bck)
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

func (r *runCmd) runServer(reg lobby.Registry) (*sync.WaitGroup, lobby.Server, error) {
	var wg sync.WaitGroup

	// listening on unix socket
	l, err := net.Listen("unix", path.Join(r.app.SocketDir, "lobby.sock"))
	if err != nil {
		return nil, nil, err
	}

	srv := rpc.NewServer(rpc.WithBucketService(reg), rpc.WithRegistryService(reg))

	wg.Add(1)
	go func(l net.Listener, srv lobby.Server) {
		defer wg.Done()
		fmt.Fprintf(r.app.out, "Listening %s requests on %s.\n", srv.Name(), l.Addr().String())
		srv.Serve(l)
	}(l, srv)

	return &wg, srv, nil
}
