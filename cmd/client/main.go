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

const ConfigIP = "http://localhost:8080"

var ring *cache_ring.CacheRing

var ringInitialized bool
var ringMutex sync.Mutex
var ringCv = sync.NewCond(&ringMutex)

func main() {
	log.SetFlags(log.Ltime)
	port := os.Args[1] // port number to listen on

	log.Printf("Started up cache client on port %s\n", port)

	ring = cache_ring.NewRing() // thread safe ring for consistent hashing

	ringInitialized = false

	go getIps() // goroutine to get the ips, runs continously

	// have to wait for ring to be initialized before we can make requests to caches
	ringCv.L.Lock()
	for !ringInitialized {
		ringCv.Wait()
	}
	ringCv.L.Unlock()

	// HTTP server for receiving commands
	http.HandleFunc("/command", handleCommand)

	http.ListenAndServe(":"+port, nil)
}

// handleCommand accepts a command via HTTP POST and routes it to the correct cache node
func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Command string `json:"command"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Command == "" {
		http.Error(w, "command required", http.StatusBadRequest)
		return
	}

	start := time.Now()

	ip := ring.FindCache(req.Command)
	log.Printf("Sending request '%s' to ip %s\n", req.Command, ip)

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Printf("Failed to connect to %s: %v\n", ip, err)
		http.Error(w, "failed to connect to cache node "+ip+": "+err.Error(), http.StatusBadGateway)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(req.Command))
	if err != nil {
		log.Printf("Failed to send to %s: %v\n", ip, err)
		http.Error(w, "failed to send command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]byte, 4096)
	n, err := conn.Read(response)
	if err != nil {
		log.Printf("Failed to read from %s: %v\n", ip, err)
		http.Error(w, "failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(start)
	log.Printf("Got response '%s' from ip %s (took %v)\n", string(response[:n]), ip, elapsed)

	json.NewEncoder(w).Encode(map[string]string{
		"target":   ip,
		"response": string(response[:n]),
		"command":  req.Command,
	})
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

		log.Printf("Updated ip list: %v\n", ips)

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
