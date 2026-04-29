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

var mu sync.Mutex

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
	/* TODO: IMPLEMENT
	- keep a timer/timestamp for each ip's last heartbeat
	- update timestamp on heartbeat
	*/
	mu.Lock()
	addr := getAddr(r)
	hit := time.Now()
	health_status_map[addr] = hit
	logger.Printf("Heartbeat from IP %s\n", addr)
	mu.Unlock()
}

// for /ips -- returns list of cache ips
func clientHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	json.NewEncoder(w).Encode(ips)
	mutex.Unlock()
	log.Printf("Sent IP list to client\n")
}

// ENTRY
func main() {
	log.SetFlags(log.Ltime)
	log.Printf("Started up config service\n")

	http.HandleFunc("/init", handleCacheStart)
	http.HandleFunc("/heartbeat", handleHeartBeat)
	http.HandleFunc("/ips", clientHandler)

	http.ListenAndServe(":8080", nil)

	go checkHealthMap()
}

func checkHealthMap() {

	/*go through map and check if any addr has a request > 10 seconds ago
	IF YES:
		1. Mark cache as dead --> tell cache che client the cache has died
		2. Cache client should update the ring
	*/
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	for addr, hit_time := range health_status_map {
		if hit_time.Before(now.Add(-10 * time.Second)) {
			for idx := range ips {
				if ips[idx] == addr {
					ips[idx] = ips[len(ips)-1] // Copy last element to index i
					ips = ips[:len(ips)-1]
					break
				}
			}
		}
	}
	time.Sleep(time.Second * 10)
}
