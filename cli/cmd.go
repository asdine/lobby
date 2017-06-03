package cli

import "github.com/spf13/cobra"

// New returns the lobby CLI application.
func New() *cobra.Command {
	a := newApp()
	a.AddCommand(newRunCmd(a))
	return a.Command
}
