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

func setCoreCmd(cmd *cobra.Command, app *app.App) {
	cmd.Short = "start a lobby server"
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
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
		case sig := <-quit:
			fmt.Println()
			app.Logger.Printf("Received %s signal. Shutting down...\n", sig)
			cancel()
			err = <-errc
		case err = <-errc:
		}

		app.Logger.Println("Shutdown complete")
		return err
	}

	cmd.Flags().StringSliceVar(&app.Config.Plugins.Backends, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringVar(&app.Config.Paths.PluginDir, "plugin-dir", "", "Location of plugins")
	cmd.Flags().IntVar(&app.Config.Grpc.Port, "grpc-port", 5656, "gRPC API port to listen on")
	cmd.Flags().IntVar(&app.Config.HTTP.Port, "http-port", 5657, "HTTP API port to listen on")
}
