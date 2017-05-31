package cli

import (
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"

	cli "gopkg.in/urfave/cli.v1"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc"
)

// NewPlugin returns a lobby plugin CLI application.
func NewPlugin(name string) *Plugin {
	plugin := Plugin{
		app: newApp(),
	}
	plugin.Name = name
	plugin.Usage = fmt.Sprintf("%s plugin", name)

	return &plugin
}

// Plugin command.
type Plugin struct {
	*app
}

// RunAsBackend runs the plugin as a lobby backend.
func (p *Plugin) RunAsBackend(fn func() (lobby.Backend, error)) error {
	p.Action = func(c *cli.Context) error {
		err := initDir(path.Join(p.DataDir, p.Name))
		if err != nil {
			return err
		}

		backend, err := fn()
		if err != nil {
			return err
		}
		defer backend.Close()

		l, err := net.Listen("unix", path.Join(p.SocketDir, fmt.Sprintf("%s.sock", p.Name)))
		if err != nil {
			return err
		}
		defer l.Close()

		srv := rpc.NewServer(rpc.WithBucketService(backend))

		return p.app.runServers(map[net.Listener]lobby.Server{
			l: srv,
		})
	}

	return p.Run(os.Args)
}

// RunAsServer runs the plugin as a lobby server.
func (p *Plugin) RunAsServer(fn func(lobby.Registry) (net.Listener, lobby.Server, error)) error {
	p.Action = func(c *cli.Context) error {
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

		return p.app.runServers(map[net.Listener]lobby.Server{
			l: srv,
		})
	}

	return p.Run(os.Args)
}
