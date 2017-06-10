package rpc

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func setFakeCommand(t *testing.T, additionalArgs ...string) func() {
	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cs = append(cs, additionalArgs...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return func() {
		execCommand = exec.Command
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	require.Equal(t, "/fake/command", cmd)
	require.Len(t, args, 2)
	require.Equal(t, "--config-dir", args[0])
	l, err := net.Listen("unix", path.Join(args[1], "sockets", "backend.sock"))
	require.NoError(t, err)
	defer l.Close()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	os.Exit(0)
}

func TestLoadBackend(t *testing.T) {
	cleanup := setFakeCommand(t)
	defer cleanup()

	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	err = os.Mkdir(path.Join(dir, "sockets"), 0755)
	require.NoError(t, err)

	bck, plg, err := LoadBackendPlugin("backend", "/fake/command", dir)
	require.NoError(t, err)
	require.Equal(t, "backend", plg.Name())
	err = bck.Close()
	require.NoError(t, err)
	err = plg.Close()
	require.NoError(t, err)
}

func TestLoadServer(t *testing.T) {
	cleanup := setFakeCommand(t)
	defer cleanup()

	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	err = os.Mkdir(path.Join(dir, "sockets"), 0755)
	require.NoError(t, err)

	plg, err := LoadPlugin("server", "/fake/command", dir)
	require.NoError(t, err)
	require.Equal(t, "server", plg.Name())
	err = plg.Close()
	require.NoError(t, err)
}
