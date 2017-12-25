package cli

import (
	"fmt"
	stdlog "log"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/asdine/lobby"
	cliapp "github.com/asdine/lobby/cli/app"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/rpc"
	"github.com/spf13/cobra"
)

// RunBackend runs a plugin as a backend.
func RunBackend(name string, fn func() (lobby.Backend, error), cfg interface{}) {
	var app cliapp.App
	root := newRootCmd(&app)
	root.Use = fmt.Sprintf("lobby-%s", name)
	root.Short = fmt.Sprintf("%s plugin", name)
	root.RunE = func(cmd *cobra.Command, args []string) error {
		var wg sync.WaitGroup

		if cfg != nil {
			if root.cfgMeta.IsDefined("plugins", "config", name) {
				err := root.cfgMeta.PrimitiveDecode(app.Config.Plugins.Config[name], cfg)
				if err != nil {
					return err
				}
			}
		}

		bck, err := fn()
		if err != nil {
			return err
		}
		defer bck.Close()

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

		l, err := net.Listen("unix", path.Join(app.Config.Paths.SocketDir, fmt.Sprintf("%s.sock", name)))
		if err != nil {
			return err
		}
		defer l.Close()

		stdlog.SetFlags(0)
		srv := rpc.NewServer(log.New(), rpc.WithTopicService(bck))

		wg.Add(1)
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

	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
