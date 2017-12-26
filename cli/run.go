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
	var pluginDir string
	var httpPort, gRPCPort int

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the lobby server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if backends != nil {
				app.Config.Plugins.Backends = backends
			}

			if pluginDir != "" {
				app.Config.Paths.PluginDir = pluginDir
			}

			if httpPort != 0 {
				app.Config.HTTP.Port = httpPort
			}

			if gRPCPort != 0 {
				app.Config.Grpc.Port = gRPCPort
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
			case sig := <-quit:
				fmt.Println()
				app.Logger.Printf("Received %s signal. Shutting down...\n", sig)
				cancel()
				err = <-errc
			case err = <-errc:
			}

			app.Logger.Println("Shutdown complete")
			return err
		},
	}

	cmd.Flags().StringSliceVar(&backends, "backend", nil, "Name of the backend to use")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Location of plugins")
	cmd.Flags().IntVar(&httpPort, "http-port", 0, "HTTP API port to listen on")
	cmd.Flags().IntVar(&gRPCPort, "grpc-port", 0, "gRPC API port to listen on")

	return &cmd
}
