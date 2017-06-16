package app

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type Paths struct {
	ConfigDir string
	PluginDir string
	SocketDir string
}

func (p *Paths) Create() error {
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