package cli

import (
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NewPlugin returns a lobby plugin CLI application.
func NewPlugin(name string) *Plugin {
	a := newApp()
	a.Command.Use = fmt.Sprintf("lobby-%s", name)
	a.Command.Short = fmt.Sprintf("%s plugin", name)
	plugin := Plugin{
		app:  a,
		name: name,
	}

	return &plugin
}

// Plugin command.
type Plugin struct {
	*app
	name string
}

// RunAsBackend runs the plugin as a lobby backend.
func (p *Plugin) RunAsBackend(fn func() (lobby.Backend, error)) error {
	p.Command.RunE = func(cmd *cobra.Command, args []string) error {
		err := initDir(path.Join(p.app.DataDir, p.name))
		if err != nil {
			return err
		}

		backend, err := fn()
		if err != nil {
			return err
		}
		defer backend.Close()

		l, err := net.Listen("unix", path.Join(p.SocketDir, fmt.Sprintf("%s.sock", p.name)))
		if err != nil {
			return err
		}
		defer l.Close()

		srv := rpc.NewServer(rpc.WithBucketService(backend))

		return runServers(ioutil.Discard, map[net.Listener]lobby.Server{
			l: srv,
		})
	}

	return p.Command.Execute()
}

// RunAsServer runs the plugin as a lobby server.
func (p *Plugin) RunAsServer(fn func(lobby.Registry) (net.Listener, lobby.Server, error)) error {
	p.Command.RunE = func(cmd *cobra.Command, args []string) error {
		conn, err := grpc.Dial("",
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return net.DialTimeout("unix", path.Join(p.SocketDir, "lobby.sock"), timeout)
			}),
		)
		if err != nil {
			return err
		}
		reg, err := rpc.NewRegistry(conn)
		if err != nil {
			return err
		}

		l, srv, err := fn(reg)
		if err != nil {
			return err
		}
		defer l.Close()

		return runServers(p.out, map[net.Listener]lobby.Server{
			l: srv,
		})
	}

	return p.Command.Execute()
}
