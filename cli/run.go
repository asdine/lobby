package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/asdine/lobby/app"
	"github.com/spf13/cobra"
)

func newRunCmd(app *app.App) *cobra.Command {
	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the lobby server",
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

	cmd.Flags().StringSliceVar(&app.Options.Plugins.Backend, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringSliceVar(&app.Options.Plugins.Server, "server", nil, "Name of the server to run")

	return &cmd
}
