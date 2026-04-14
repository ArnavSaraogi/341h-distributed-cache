package main

import (
	"distributedCache/cache_ring"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// logger
var logger = log.New(os.Stderr, "[CACHE CLIENT]: ", log.Ltime)

const ConfigIP = "http://localhost:8080"

var ring *cache_ring.CacheRing

var ringInitialized bool
var ringMutex sync.Mutex
var ringCv = sync.NewCond(&ringMutex)

func main() {
	logger.Printf("Started up cache client\n")

	ring = cache_ring.NewRing() // thread safe ring for consistent hashing

	ringInitialized = false

	go getIps() // goroutine to get the ips, runs continously

	// have to wait for ring to be initialized before we can make requests to caches
	ringCv.L.Lock()
	for !ringInitialized {
		ringCv.Wait()
	}
	ringCv.L.Unlock()

	ring.PrintCachesIPsAndHashes()

	// requests
	requests := []string{"GET jayleen",
		"GET sanjiv",
		"GET jesse",
		"GET aisha",
		"GET amara",
		"GET andre",
		"GET arnav",
		"GET mikey",
		"GET yuki",
	}

	// sending requests to respective caches
	for _, request := range requests {
		start := time.Now()

		ip := ring.FindCache(request)
		logger.Printf("Sending request '%s' to ip %s\n", request, ip)

		conn, err := net.Dial("tcp", ip)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = conn.Write([]byte(request))
		if err != nil {
			log.Fatalln(err)
		}

		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Printf("Got response '%s' from ip %s (took %v)\n", string(response[:n]), ip, time.Since(start))

		conn.Close()
	}
}

// thread that periodically gets IP list from config service
func getIps() {

	// function to fetch the ips
	fetch := func() {
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

		ringCv.L.Lock()
		ringInitialized = true
		ringCv.Broadcast()
		ringCv.L.Unlock()
	}

	fetch() // fetch ips immediately

	// periodic loop that fetches ips every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		fetch()
	}
}
