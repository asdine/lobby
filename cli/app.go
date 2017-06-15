package cli

import (
	"fmt"
	"log"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newApp() *app {
	var a app

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	defaultConfigDir := path.Join(home, ".config/lobby")

	a.PluginDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cmd := cobra.Command{
		Use:          "lobby",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			a.DataDir = path.Join(defaultConfigDir, "data")
			a.SocketDir = path.Join(defaultConfigDir, "sockets")
			return a.init()
		},
	}

	cmd.PersistentFlags().StringVarP(&a.ConfigDir, "config-dir", "c", defaultConfigDir, "Location of Lobby configuration files")
	cmd.PersistentFlags().StringVar(&a.PluginDir, "plugin-dir", ".", "Location of plugins")

	a.Command = &cmd
	return &a
}

type app struct {
	*cobra.Command

	ConfigDir string
	DataDir   string
	SocketDir string
	PluginDir string
}

func (a *app) init() error {
	return a.initDirectories()
}

func (a *app) initDirectories() error {
	paths := []string{
		a.ConfigDir,
		a.DataDir,
		a.SocketDir,
	}

	for _, path := range paths {
		err := initDir(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func initDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "Can't create directory %s", path)
		}

		return nil
	}

	if !fi.Mode().IsDir() {
		return fmt.Errorf("'%s' is not a valid directory", path)
	}

	return nil
}
