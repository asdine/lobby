package main

import (
	"log"
	"net"

	"github.com/asdine/lobby/bolt"
	"github.com/asdine/lobby/rpc"
)

func main() {
	reg, err := bolt.NewRegistry("blabla.db")
	if err != nil {
		log.Fatal(err)
	}
	defer reg.Close()

	bck, err := bolt.NewBackend("backend.db")
	if err != nil {
		log.Fatal(err)
	}
	defer bck.Close()

	reg.RegisterBackend("bolt", bck)

	l, err := net.Listen("tcp", ":5658")
	if err != nil {
		log.Fatal(err)
	}

	server := rpc.NewServer(reg)
	if err := server.Serve(l); err != nil {
		log.Fatal(err)
	}
}
