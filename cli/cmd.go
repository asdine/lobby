package cli

import cli "gopkg.in/urfave/cli.v1"

// New returns the lobby CLI application.
func New() *cli.App {
	a := newApp()
	a.App.Commands = []cli.Command{
		newRunCmd(a),
	}
	return a.App
}
