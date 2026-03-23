package main

import (
	"distributedCache/node"
	"log"
	"net"
	"os"
)

const Capacity = 100

func main() {

	port_num := os.Args[1] // port number to listen to will be a command line arg
	socket := "localhost:" + port_num

	node := node.NewNode(Capacity) // initialize new cache node

	ln, err := net.Listen("tcp", socket)
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()
	for {
		// wait for incoming client connections
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// handle the connection
		go node.HandleConnection(conn)
	}
}
