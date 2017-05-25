package cli

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

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

		return p.backendAction(backend)
	}

	return p.Run(os.Args)
}

func (p *Plugin) backendAction(backend lobby.Backend) error {
	var wg sync.WaitGroup

	l, err := net.Listen("unix", path.Join(p.SocketDir, fmt.Sprintf("%s.sock", p.Name)))
	if err != nil {
		return err
	}
	defer l.Close()

	server := rpc.NewServer(rpc.WithBucketService(backend))

	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Serve(l)
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Fprintf(p.out, "\nStopping servers...")
	err = server.Stop()
	if err != nil {
		fmt.Fprintf(p.out, " Error: %s\n", err.Error())
	} else {
		fmt.Fprintf(p.out, " OK\n")
	}

	wg.Wait()
	return nil
}
