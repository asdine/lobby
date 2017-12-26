package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/asdine/lobby"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var execCommand = exec.Command

type process struct {
	*os.Process
	m      sync.Mutex
	conn   *grpc.ClientConn
	name   string
	closed bool
}

func (p *process) Name() string {
	return p.name
}

func (p *process) Wait() error {
	if p.closed {
		return nil
	}

	status, err := p.Process.Wait()
	if err != nil {
		return err
	}

	p.m.Lock()
	defer p.m.Unlock()

	if !p.closed {
		p.closed = true
		return fmt.Errorf("plugin %s exited unexpectedly", p.name)
	}

	if !status.Success() {
		return fmt.Errorf("plugin %s crashed during exit", p.name)
	}

	return nil
}

func (p *process) Close() error {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			return err
		}
		p.conn = nil
	}

	return p.Signal(syscall.SIGTERM)
}

// LoadPlugin loads a plugin.
func LoadPlugin(ctx context.Context, name, cmdPath, dataDir, configFile string) (lobby.Plugin, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	args := []string{
		"--data-dir", dataDir,
	}

	if configFile != "" {
		args = append(args, "-c", configFile)
	}

	cmd := execCommand(cmdPath, args...)
	prefixFn := func() []byte {
		return []byte(time.Now().Format("2006/01/02 15:04:05") + " i | " + name + ": ")
	}

	cmd.Stdout = lobby.NewFuncPrefixWriter(prefixFn, os.Stdout)
	cmd.Stderr = lobby.NewFuncPrefixWriter(prefixFn, os.Stderr)
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
func LoadBackendPlugin(ctx context.Context, name, cmdPath, dataDir, configFile string) (lobby.Backend, lobby.Plugin, error) {
	plugin, err := LoadPlugin(ctx, name, cmdPath, dataDir, configFile)
	if err != nil {
		return nil, nil, err
	}

	socketPath := path.Join(dataDir, "sockets", fmt.Sprintf("%s.sock", name))
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

Loop:
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
				break Loop
			}
		case <-ctx.Done():
			err := plugin.Close()
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to kill process %s", name)
			}

			return nil, nil, ctx.Err()
		}
	}

	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(1*time.Second),
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
