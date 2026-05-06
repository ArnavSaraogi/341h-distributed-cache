package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ip checker maps ip -> last seen timestamp
var health_status_map = map[string]time.Time{}

// logger
var logger = log.New(os.Stderr, "[CONFIG SERVICE]: ", log.Ltime)

// HELPERS
func getAddr(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr) // IP is correct, port is not
	body, _ := io.ReadAll(r.Body)
	port := strings.TrimSpace(string(body)) // port from the cache server itself

	addr := "[" + host + "]" + ":" + port // HARDCODED BRACKETS SINCE ON SAME MACHINE

	return addr
}

// HANDLERS
// for /init -- adds an ip to its list of known ips
func handleCacheStart(w http.ResponseWriter, r *http.Request) {
	addr := getAddr(r)

	mutex.Lock()
	ips = append(ips, addr)
	mutex.Unlock()

	log.Printf("Added IP %s in cache IP list\n", addr)
}

var ips []string
var mutex sync.Mutex

// for /heartbeat -- updates heartbeat timestamp
func handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	addr := getAddr(r)
	health_status_map[addr] = time.Now()
	logger.Printf("Heartbeat from IP %s\n", addr)
	mutex.Unlock()
}

// for /ips -- returns list of cache ips
func clientHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	json.NewEncoder(w).Encode(ips)
	log.Printf("Sent IP list to client\n")
}

// ENTRY
func main() {
	log.SetFlags(log.Ltime)
	log.Printf("Started up config service\n")

	go checkHealthMap()

	http.HandleFunc("/init", handleCacheStart)
	http.HandleFunc("/heartbeat", handleHeartBeat)
	http.HandleFunc("/ips", clientHandler)

	http.ListenAndServe(":8080", nil)
}

const heartbeatTimeout = 10 * time.Second
const checkInterval = 2500 * time.Millisecond

func checkHealthMap() {
	for {
		time.Sleep(checkInterval)

		mutex.Lock()
		now := time.Now()
		for addr, hitTime := range health_status_map {
			if hitTime.Before(now.Add(-heartbeatTimeout)) {
				// remove from ips list
				for idx := range ips {
					if ips[idx] == addr {
						ips[idx] = ips[len(ips)-1]
						ips = ips[:len(ips)-1]
						break
					}
				}
				delete(health_status_map, addr)
				logger.Printf("Removed dead cache node %s\n", addr)
			}
		}
		mutex.Unlock()
	}
}
