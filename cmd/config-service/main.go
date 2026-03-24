package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
)

/*
also endpoint for starting up a cache
*/

/*
1. set up http endpoint that cache server function can make post reqeust to
2. keep track of active ips from the information hitting the http endpoint
3. send the list via an async function to client
*/

var ips []string
var mutex sync.Mutex

// heartbeating
func cacheHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	mutex.Lock()
	ips = append(ips, ip)
	mutex.Unlock()
}

func main() {
	http.HandleFunc("/ip_addresses", cacheHandler)
	http.HandleFunc("/ips", clientHandler)
	http.ListenAndServe(":8080", nil)

}
func clientHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	json.NewEncoder(w).Encode(ips)
	mutex.Unlock()
	//comes back to client as ["1.2.3.4", "5.6.7.8"]
}
