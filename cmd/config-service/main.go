package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

// logger
var logger = log.New(os.Stderr, "[CONFIG SERVICE]: ", log.Ltime)

// HANDLERS
// for /init -- adds an ip to its list of known ips
func handleCacheStart(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	mutex.Lock()
	ips = append(ips, ip)
	mutex.Unlock()

	logger.Printf("Added IP %s in cache IP list\n", ip)
}

var ips []string
var mutex sync.Mutex

// for /heartbeat -- updates heartbeat timestamp
func handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	/* TODO: IMPLEMENT
	- keep a timer/timestamp for each ip's last heartbeat
	- update timestamp on heartbeat
	*/
}

// for /ips -- returns list of cache ips
func clientHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	json.NewEncoder(w).Encode(ips)
	mutex.Unlock()
}

// ENTRY
func main() {
	logger.Printf("Started up config service\n")

	http.HandleFunc("/init", handleCacheStart)
	http.HandleFunc("/heartbeat", handleHeartBeat)
	http.HandleFunc("/ips", clientHandler)

	http.ListenAndServe(":8080", nil)
}
