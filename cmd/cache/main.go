package main

import (
	"distributedCache/node"
	"log"
	"net"
)

const Capacity = 100

func main() {
	node := node.NewNode(Capacity)
	ln, err := net.Listen("tcp", "localhost::8080")
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
