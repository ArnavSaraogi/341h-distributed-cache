package main

import (
	"distributedCache/node"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
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
	//node gets init in config
	_, err = http.Post("http://localhost:8080/init", "text/plain", nil)
	if err != nil {
		panic(err)
	}

	//hit endpoint every 10 seconds ---> health update
	go sendIp()
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

func sendIp() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_, err := http.Post("http://localhost:8080/heartbeat", "text/plain", nil)
		if err != nil {
			panic(err)
		}
	}
}
