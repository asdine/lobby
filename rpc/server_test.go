package rpc_test

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/rpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func newServer(t *testing.T, r lobby.Registry) (*grpc.ClientConn, func()) {
	dir, err := ioutil.TempDir("", "lobby")
	require.NoError(t, err)

	socketPath := path.Join(dir, "lobby.sock")
	l, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	srv := rpc.NewServer(rpc.WithBucketService(r), rpc.WithRegistryService(r))

	go func() {
		srv.Serve(l)
	}()

	conn, err := grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		srv.Stop()
		os.RemoveAll(dir)
	}
}
