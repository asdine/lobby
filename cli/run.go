package cli

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/http"
	"github.com/asdine/lobby/rpc"
	"github.com/spf13/cobra"
)

func newRunCmd(a *app) *cobra.Command {
	r := runCmd{
		app:    a,
		stdout: log.New(os.Stderr, "[lobby] ", 0),
		stderr: log.New(os.Stderr, "[lobby] ", log.LstdFlags),
	}

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
	stdout      *log.Logger
	stderr      *log.Logger
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

	quit, err := r.runAllServers(reg)
	if err != nil {
		return err
	}

	err = r.loadPlugins(reg)
	if err != nil {
		r.closeAll(quit)
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println()

	r.closeAll(quit)
	return nil
}

func (r *runCmd) closeAll(quit chan struct{}) {
	r.stdout.Printf("Shutting down plugins")
	if err := r.closePlugins(); err != nil {
		r.stdout.Printf("Error while stopping plugins: %s\n", err.Error())
	}

	r.stdout.Printf("Shutting down Lobby")
	// Sending message to stop servers
	quit <- struct{}{}
	// Waiting for servers to stop
	<-quit
	r.stdout.Printf("Shutting down complete\n")
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

func (r *runCmd) runAllServers(reg lobby.Registry) (chan struct{}, error) {
	// gRPC: listening on unix socket
	lsock, err := net.Listen("unix", path.Join(r.app.SocketDir, "lobby.sock"))
	if err != nil {
		return nil, err
	}

	// gRPC: listening on port
	// tmp: harcoded port
	lgRPC, err := net.Listen("tcp", ":5656")
	if err != nil {
		return nil, err
	}

	// http: listening on port
	// tmp: harcoded port
	lhttp, err := net.Listen("tcp", ":5657")
	if err != nil {
		return nil, err
	}

	quit := r.runServers(map[net.Listener]lobby.Server{
		lsock: rpc.NewServer(rpc.WithBucketService(reg), rpc.WithRegistryService(reg)),
		lgRPC: rpc.NewServer(rpc.WithBucketService(reg), rpc.WithRegistryService(reg)),
		lhttp: http.NewServer(http.NewHandler(reg, log.New(os.Stderr, "[http] ", log.LstdFlags))),
	})

	return quit, nil
}

func (r *runCmd) runServers(servers map[net.Listener]lobby.Server) chan struct{} {
	var wg sync.WaitGroup
	quit := make(chan struct{})

	for l, srv := range servers {
		wg.Add(1)
		go func(l net.Listener, srv lobby.Server) {
			defer wg.Done()
			log.New(os.Stderr, fmt.Sprintf("[%s] ", srv.Name()), 0).Printf("Listening %s requests on %s.\n", srv.Name(), l.Addr().String())
			_ = srv.Serve(l)
		}(l, srv)
	}

	go func() {
		<-quit

		for _, srv := range servers {
			err := srv.Stop()
			if err != nil {
				// TODO: return all errors in a separate chan
				r.stderr.Print(err)
			}
		}

		wg.Wait()
		quit <- struct{}{}
	}()

	return quit
}
