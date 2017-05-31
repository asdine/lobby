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

type Plugin interface {
	Name() string
	Close() error
}

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
			time.Sleep(100 * time.Millisecond)
			return net.DialTimeout("unix", b.socketPath, timeout)
		}),
	)
	if err != nil {
		return nil, err
	}

	return rpc.NewBackend(conn)
}

func LoadBackend(name, cmdPath, configDir string) (Backend, error) {
	cmd := exec.Command(cmdPath, "--config-dir", configDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return &backend{
		socketPath: path.Join(configDir, "sockets", fmt.Sprintf("%s.sock", name)),
		plugin: &plugin{
			name:    name,
			process: cmd.Process,
		},
	}, nil
}

func LoadServer(name, cmdPath, configDir string) (Plugin, error) {
	cmd := exec.Command(cmdPath, "--config-dir", configDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return &plugin{
		process: cmd.Process,
	}, nil
}
