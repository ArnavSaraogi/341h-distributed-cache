package main

import (
	"distributedCache/cache_ring"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

var ring *cache_ring.CacheRing

func main() {
	ring = cache_ring.NewRing() // thread safe ring for consistent hashing

	go getIps() // goroutine to get the ips, runs continously

	request := "GET sanjiv"

	ip := ring.FindCache(request)

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Fatalln(err)
	}
}

// thread that periodically gets IP list from config service
func getIps() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ips := []string{}

		// get updated list from config service
		res, err := http.Get("http://localhost:8080/ips") // storing config service on same machine for now
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
	}
}
