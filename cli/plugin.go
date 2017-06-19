package cli

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/cli/app"
	"github.com/asdine/lobby/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// RunPlugin runs a plugin as a standalone application.
func RunPlugin(name string, startFn func(lobby.Registry) error, stopFn func() error) error {
	app := app.NewApp()
	cmd := newRootCmd(app)

	cmd.Use = fmt.Sprintf("lobby-%s", name)
	cmd.Short = fmt.Sprintf("%s plugin", name)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var wg sync.WaitGroup

		conn, err := grpc.Dial("",
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return net.DialTimeout("unix", path.Join(app.Config.Paths.SocketDir, "lobby.sock"), timeout)
			}),
		)
		if err != nil {
			return err
		}
		reg, err := rpc.NewRegistry(conn)
		if err != nil {
			return err
		}

		go func() {
			defer wg.Done()
			err := startFn(reg)
			if err != nil {
				log.Fatal(err)
			}
		}()

		ch := make(chan os.Signal, 1)

		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		<-ch
		err = stopFn()
		if err != nil {
			return err
		}

		wg.Wait()
		return nil
	}
	return cmd.Execute()
}

// RunBackend runs a plugin as a backend.
func RunBackend(name string, bck lobby.Backend) error {
	app := app.NewApp()
	cmd := newRootCmd(app)
	cmd.Use = fmt.Sprintf("lobby-%s", name)
	cmd.Short = fmt.Sprintf("%s plugin", name)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var wg sync.WaitGroup
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		defer bck.Close()

		l, err := net.Listen("unix", path.Join(app.Config.Paths.SocketDir, fmt.Sprintf("%s.sock", name)))
		if err != nil {
			return err
		}
		defer l.Close()

		srv := rpc.NewServer(rpc.WithBucketService(bck))

		go func() {
			defer wg.Done()
			_ = srv.Serve(l)
		}()

		<-ch
		err = srv.Stop()
		if err != nil {
			return err
		}

		wg.Wait()
		return nil
	}

	return cmd.Execute()
}
