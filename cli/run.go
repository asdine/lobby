package cli

import (
	"io/ioutil"
	"log"
	"net"
	"path"

	"google.golang.org/grpc/grpclog"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/rpc"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	defaultAddr = ":5656"
)

func newRunCmd(a *app) cli.Command {
	cmd := runCmd{
		app: a,
	}

	return cli.Command{
		Name:   "run",
		Usage:  "Run Lobby",
		Action: cmd.run,
	}
}

type runCmd struct {
	app     *app
	mainSrv lobby.Server
}

func (s *runCmd) run(_ *cli.Context) error {
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))

	return s.runMainServer()
}

func (s *runCmd) runMainServer() error {
	dataPath := path.Join(s.app.DataDir, "bolt")
	registryPath := path.Join(dataPath, "registry.db")

	err := initDir(dataPath)
	if err != nil {
		return err
	}

	reg, err := bolt.NewRegistry(registryPath)
	if err != nil {
		return err
	}

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
	defer l.Close()

	// listening on unix socket
	lsock, err := net.Listen("unix", path.Join(s.app.SocketDir, "lobby.sock"))
	if err != nil {
		return err
	}
	defer lsock.Close()

	return s.app.runServers(map[net.Listener]lobby.Server{
		l:     srv,
		lsock: srv,
	})
}
