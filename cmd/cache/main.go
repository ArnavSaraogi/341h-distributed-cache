package main

import (
	"distributedCache/node"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const Capacity = 100
const ConfigIP = "http://localhost:8080"

func main() {
	log.SetFlags(log.Ltime)

	port := os.Args[1] // port number to listen to will be a command line arg

	log.Printf("Started up cache server with port %s\n", port)

	node := node.NewNode(Capacity) // initialize new cache node

	// node gets init in config
	_, err := http.Post(ConfigIP+"/init", "text/plain", strings.NewReader(port))
	if err != nil {
		panic(err)
	}

	// hearbeating
	go heartbeat(port)

	// listents to tcp request on this specific port
	ln, err := net.Listen("tcp", ":"+port)
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

// heartbeating to config service
func heartbeat(port string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_, err := http.Post(ConfigIP+"/heartbeat", "text/plain", strings.NewReader(port))
		if err != nil {
			panic(err)
		}
		// log.Printf("sent heartbeat from port %s", port)
	}
}
