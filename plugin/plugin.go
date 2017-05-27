package plugin

import (
	"bytes"
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
	Backend() (lobby.Backend, error)
	Close() error
}

func Load(name, cmdPath, configDir, dataDir string) (Plugin, error) {
	cmd := exec.Command(cmdPath, "--config-dir", configDir, "--data-dir", dataDir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return &plugin{
		process:    cmd.Process,
		socketPath: path.Join(configDir, "sockets"),
	}, nil
}

type plugin struct {
	process    *os.Process
	socketPath string
}

func (p *plugin) Backend() (lobby.Backend, error) {
	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			sock, err := net.DialTimeout("unix", p.socketPath, timeout)
			return sock, err
		}),
	)
	if err != nil {
		return nil, err
	}

	return rpc.NewBackend(conn)
}

func (p *plugin) Close() error {
	return p.process.Kill()
}
