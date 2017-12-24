package app

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config of the application.
type Config struct {
	Paths   Paths
	Plugins Plugins
}

// Plugins contains the list of backend and server plugins.
type Plugins struct {
	Backend []string
	Config  map[string]toml.Primitive
}

// Paths contains directory paths needed by the app.
type Paths struct {
	ConfigDir string `toml:"-"`
	PluginDir string `toml:"plugin-dir"`
	SocketDir string `toml:"socket-dir"`
}

// Create the ConfigDir and SocketDir if they don't exist.
func (p *Paths) Create() error {
	if p.ConfigDir == "" {
		return errors.New("unspecified config directory")
	}

	if p.SocketDir == "" {
		return errors.New("unspecified socket directory")
	}

	paths := []string{
		p.ConfigDir,
		p.SocketDir,
	}

	for _, path := range paths {
		err := createDir(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func createDir(path string) error {
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
