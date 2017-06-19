package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/asdine/lobby/cli/app"
	"github.com/spf13/cobra"
)

func newRunCmd(app *app.App) *cobra.Command {
	var backends []string
	var servers []string
	var pluginDir string

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the lobby server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if backends != nil {
				app.Config.Plugins.Backend = backends
			}

			if servers != nil {
				app.Config.Plugins.Server = servers
			}

			if pluginDir != "" {
				app.Config.Paths.PluginDir = pluginDir
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			errc := make(chan error)
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

			go func() {
				errc <- app.Run(ctx)
			}()

			var err error

			select {
			case <-quit:
				fmt.Println()
				cancel()
				err = <-errc
			case err = <-errc:
			}

			return err
		},
	}

	cmd.Flags().StringSliceVar(&backends, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringSliceVar(&servers, "server", nil, "Name of the server to run")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Location of plugins")

	return &cmd
}
