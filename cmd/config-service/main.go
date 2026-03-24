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
// for /init
func handleCacheStart(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	mutex.Lock()
	ips = append(ips, ip)
	mutex.Unlock()
}

var ips []string
var mutex sync.Mutex

// heartbeating
func handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	/* TODO: IMPLEMENT
	- keep a timer/timestamp for each ip's last heartbeat
	- update timestamp on heartbeat
	*/
}

func main() {
	logger.Println("Started up config service")

	http.HandleFunc("/init", handleCacheStart)
	http.HandleFunc("/heartbeat", handleHeartBeat)
	http.HandleFunc("/ips", clientHandler)
	http.ListenAndServe(":8080", nil)
}

// comes back to client as ["1.2.3.4", "5.6.7.8"]
func clientHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	json.NewEncoder(w).Encode(ips)
	mutex.Unlock()
}
