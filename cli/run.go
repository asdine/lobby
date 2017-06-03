package cli

import (
	"net"
	"path"

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
	}

	return &cmd
}

type runCmd struct {
	app     *app
	mainSrv lobby.Server
}

func (s *runCmd) run(cmd *cobra.Command, args []string) error {
	return s.runMainServer()
}

func (s *runCmd) runMainServer() error {
	err := s.app.loadBackendPlugins()
	if err != nil {
		return err
	}

	dataPath := path.Join(s.app.DataDir, "bolt")
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
	for _, p := range s.app.Backends {
		bck, err := p.Backend()
		if err != nil {
			return err
		}

		reg.RegisterBackend(p.Name(), bck)
	}

	srv := rpc.NewServer(rpc.WithBucketService(reg), rpc.WithRegistryService(reg))

	// listening on specific port
	l, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		return err
	}

	// listening on unix socket
	lsock, err := net.Listen("unix", path.Join(s.app.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}

	err = s.app.loadServerPlugins()
	if err != nil {
		return err
	}

	return s.app.runServers(map[net.Listener]lobby.Server{
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
