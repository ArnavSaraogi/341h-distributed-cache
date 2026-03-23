package main

import (
	"distributedCache/cache_ring"
	"log"
	"net"
)

func main() {
	// TODO: GET LIST OF CACHES FROM CONFIG SERVICE
	ips := []string{}

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
