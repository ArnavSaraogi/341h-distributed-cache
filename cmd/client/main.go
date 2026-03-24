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
	ring = cache_ring.NewRing()

	go getIps() // goroutine to get the ips, runs continously

	request := "GET sanjiv"
	/*
		1. getc correct cache ip for incoming req
		2. open connection to cache -> pass in cache ip as param
		3. interact with cache
	*/
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

		// add caches to the ring
		for i := 0; i < len(ips); i++ {
			ring.AddIP((ips)[i])
		}
	}
}
