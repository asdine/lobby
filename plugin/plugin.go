package plugin

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	"google.golang.org/grpc"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc"
)

var execCommand = exec.Command

// Plugin is a generic lobby plugin.
type Plugin interface {
	Name() string
	Close() error
}

// Backend is a backend plugin.
type Backend interface {
	Plugin

	Backend() (lobby.Backend, error)
}

type plugin struct {
	name    string
	process *os.Process
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Close() error {
	return p.process.Kill()
}

type backend struct {
	*plugin
	socketPath string
}

func (b *backend) Backend() (lobby.Backend, error) {
	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", b.socketPath, timeout)
		}),
	)
	if err != nil {
		return nil, err
	}

	return rpc.NewBackend(conn)
}

// LoadBackend loads a backend plugin.
func LoadBackend(name, cmdPath, configDir string) (Backend, error) {
	cmd := execCommand(cmdPath, "--config-dir", configDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	socketPath := path.Join(configDir, "sockets", fmt.Sprintf("%s.sock", name))
	var i int
	for i < 5 {
		if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
			break
		}

		i++
		time.Sleep(10 * time.Millisecond)
	}

	return &backend{
		socketPath: socketPath,
		plugin: &plugin{
			name:    name,
			process: cmd.Process,
		},
	}, nil
}

// LoadServer loads a server plugin.
func LoadServer(name, cmdPath, configDir string) (Plugin, error) {
	cmd := execCommand(cmdPath, "--config-dir", configDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return &plugin{
		name:    name,
		process: cmd.Process,
	}, nil
}
