package rpc

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/asdine/lobby"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var execCommand = exec.Command

type process struct {
	*os.Process
	conn *grpc.ClientConn
	name string
}

func (p *process) Name() string {
	return p.name
}

func (p *process) Close() error {
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			return err
		}
	}

	go func() {
		p.Signal(syscall.SIGTERM)
	}()
	_, err := p.Wait()
	return err
}

// LoadPlugin loads a plugin.
func LoadPlugin(name, cmdPath, configDir string) (lobby.Plugin, error) {
	cmd := execCommand(cmdPath, "--config-dir", configDir)
	prefix := fmt.Sprintf("[%s]\t", name)
	cmd.Stdout = lobby.NewPrefixWriter(prefix, os.Stdout)
	cmd.Stderr = lobby.NewPrefixWriter(prefix, os.Stderr)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return &process{
		Process: cmd.Process,
		name:    name,
	}, nil
}

// LoadBackendPlugin loads a backend plugin.
func LoadBackendPlugin(name, cmdPath, configDir string) (lobby.Backend, lobby.Plugin, error) {
	plugin, err := LoadPlugin(name, cmdPath, configDir)
	if err != nil {
		return nil, nil, err
	}

	socketPath := path.Join(configDir, "sockets", fmt.Sprintf("%s.sock", name))
	c := time.Tick(10 * time.Millisecond)
	timeout := time.After(5 * time.Second)

Loop:
	for {
		select {
		case <-c:
			if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
				break Loop
			}
		case <-timeout:
			err := plugin.Close()
			if err != nil {
				return nil, nil, errors.Wrapf(err, "plugin %s load time exceeded: failed to kill process", name)
			}

			return nil, nil, errors.Errorf("plugin %s load time exceeded", name)
		}
	}

	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	plugin.(*process).conn = conn
	bck, err := NewBackend(conn)
	if err != nil {
		return nil, nil, err
	}

	return bck, plugin, nil
}
