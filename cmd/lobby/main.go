package main

import (
	"os"

	"github.com/asdine/lobby/cli"
)

func main() {
	err := cli.New().Execute()
	if err != nil {
		os.Exit(1)
	}
}
