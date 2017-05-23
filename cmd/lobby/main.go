package main

import (
	"log"
	"os"

	"github.com/asdine/lobby/cli"
)

func main() {
	cmd := cli.New()

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
