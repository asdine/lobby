package cli

import (
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/asdine/lobby/cli/app"
	"github.com/spf13/cobra"
)

// New returns the lobby CLI application.
func New() *cobra.Command {
	var app app.App
	cmd := newRootCmd(&app)
	setCoreCmd(cmd.Command, &app)
	return cmd.Command
}

func newRootCmd(app *app.App) *rootCmd {
	var configPath string
	var cfgMeta toml.MetaData

	cmd := cobra.Command{
		Use:          "lobby",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(configPath)
			if err == nil {
				cfgMeta, err = toml.DecodeReader(f, &app.Config)
				_ = f.Close()
				if err != nil {
					return err
				}
			}

			if app.Config.Paths.SocketDir == "" {
				app.Config.Paths.SocketDir = path.Join(app.Config.Paths.DataDir, "sockets")
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config-file", "c", "./lobby.toml", "Path to the Lobby config file")
	cmd.PersistentFlags().StringVar(&app.Config.Paths.DataDir, "data-dir", ".lobby", "Path to Lobby data files")
	cmd.PersistentFlags().BoolVar(&app.Config.Debug, "debug", false, "Enable debug mode")

	return &rootCmd{
		Command: &cmd,
		cfgMeta: &cfgMeta,
	}
}

type rootCmd struct {
	*cobra.Command
	cfgMeta *toml.MetaData
}
