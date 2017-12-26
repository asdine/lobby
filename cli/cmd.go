package cli

import (
	"os"

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
	var cfgMeta toml.MetaData

	cmd := cobra.Command{
		Use:          "lobby",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if app.ConfigPath == "" {
				return nil
			}

			f, err := os.Open(app.ConfigPath)
			if err != nil {
				return err
			}
			defer f.Close()

			cfgMeta, err = toml.DecodeReader(f, &app.Config)
			return err
		},
	}

	cmd.PersistentFlags().StringVarP(&app.ConfigPath, "config-file", "c", "", "Path to the Lobby config file")
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
