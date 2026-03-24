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

func main() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	var ips []string

	go func() {
		for range ticker.C {
			getIps(&ips)
		}
	}()

	// add caches to the ring
	ring := cache_ring.NewRing()
	for i := 0; i < len(ips); i++ {
		ring.AddIP(ips[i])
	}
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

func getIps(ips *[]string) {
	res, err := http.Get("http://localhost:8080/ips") // storing config service on same machine for now
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, ips)
	if err != nil {
		panic(err)
	}
}
