package main

import (
	"distributedCache/node"
	"fmt"
	"log"
	"net"
	"os"
)

const Capacity = 100

/*
 */
func main() {

	port_num := os.Args[1] // port number to listen to will be a command line arg
	socket := ":" + port_num

	fmt.Printf("%s\n", socket)

	node := node.NewNode(Capacity) // initialize new cache node

	// this node listents to tcp request on this specific socket
	ln, err := net.Listen("tcp", socket)

	if err != nil {
		log.Fatalln(err)
	}
	/*
		1. make an async function that sends socket to an endpoint in config service
		via a post request (request body is socket)
	*/
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
