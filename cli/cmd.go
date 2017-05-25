package cli

import (
	cli "gopkg.in/urfave/cli.v1"
)

// New returns the lobby CLI application.
func New() *cli.App {
	a := newApp()
	return a.App
}
