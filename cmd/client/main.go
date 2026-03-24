package main

import (
	"distributedCache/cache_ring"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// logger
var logger = log.New(os.Stderr, "[CACHE CLIENT]: ", log.Ltime)

const ConfigIP = "http://localhost:8080"

var ring *cache_ring.CacheRing

func main() {
	logger.Printf("Started up cache client\n")

	ring = cache_ring.NewRing() // thread safe ring for consistent hashing

	go getIps() // goroutine to get the ips, runs continously

	request := "GET sanjiv"
	ip := ring.FindCache(request)

	logger.Printf("Sending request '%s' to ip %s\n", request, ip)

	// start connection with cache
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Fatalln(err)
	}

	// send request through connection
	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Fatalln(err)
	}

	// get response from cache
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(response[:n]))

	logger.Printf("Got response '%s' from ip %s\n", response, ip)
}

// thread that periodically gets IP list from config service
func getIps() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ips := []string{}

		// get updated list from config service
		res, err := http.Get(ConfigIP + "/ips") // storing config service on same machine for now
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(body, &ips) // storing returned list in ips
		if err != nil {
			panic(err)
		}

		// add caches to the ring; O(n^2) :(
		ring.ClearRing()
		for i := 0; i < len(ips); i++ {
			ring.AddIP((ips)[i])
		}

		logger.Printf("Updated ip list: %v\n", ips)
	}
}
