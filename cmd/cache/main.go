package main

import (
	"distributedCache/node"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// logger
var logger = log.New(os.Stderr, "[CACHE SERVER]: ", log.Ltime)

const Capacity = 100
const ConfigIP = "http://localhost:8080"

func main() {

	port_num := os.Args[1] // port number to listen to will be a command line arg
	socket := ":" + port_num

	logger.Printf("Started up cache server %s\n", socket)

	node := node.NewNode(Capacity) // initialize new cache node

	// node gets init in config
	_, err := http.Post(ConfigIP+"/init", "text/plain", nil)
	if err != nil {
		panic(err)
	}

	// hearbeating
	go sendIp()

	// listents to tcp request on this specific socket
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

// heartbeating to config service
func sendIp() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_, err := http.Post(ConfigIP+"/heartbeat", "text/plain", nil)
		if err != nil {
			panic(err)
		}
	}
}
