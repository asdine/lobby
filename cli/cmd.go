package cli

import (
	"log"
	"path"

	"github.com/asdine/lobby/app"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// New returns the lobby CLI application.
func New() *cobra.Command {
	app := app.NewApp()
	cmd := newRootCmd(app)
	cmd.AddCommand(newRunCmd(app))
	return cmd
}

func newRootCmd(app *app.App) *cobra.Command {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	defaultConfigDir := path.Join(home, ".config/lobby")

	cmd := cobra.Command{
		Use:          "lobby",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			app.Options.Paths.SocketDir = path.Join(defaultConfigDir, "sockets")
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&app.Options.Paths.ConfigDir, "config-dir", "c", defaultConfigDir, "Location of Lobby configuration files")
	cmd.PersistentFlags().StringVar(&app.Options.Paths.PluginDir, "plugin-dir", ".", "Location of plugins")

	return &cmd
}
