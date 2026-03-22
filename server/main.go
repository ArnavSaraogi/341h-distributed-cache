package main

import (
	"distributedcache/cache"
	"log"
	"net"
	"strings"
)

// set an arbitrary capacity for now
var Cache = cache.NewCache(100)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}
		if n == 0 {
			return
		}
		req := string(buf[:n])
		req = strings.TrimSpace(req)
		parts := strings.Fields(req)
		cmd := parts[0]
		if cmd == "GET" {
			response := handleGet(parts)
			conn.Write([]byte(response))
		}
		if cmd == "PUT" {
			response := handlePut(parts)
			conn.Write([]byte(response))
		}
	}
}

func handlePut(parts []string) string {
	if len(parts) < 3 {
		return "invalid request\n"
	}
	newkey := strings.TrimSpace(parts[1])
	newval := strings.TrimSpace(strings.Join(parts[2:], " "))
	cache.CachePut(Cache, newkey, newval)
	return "ok\n"
}

func handleGet(parts []string) string {
	key := strings.TrimSpace(parts[1])
	return cache.CacheGet(Cache, key)
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()
	for {
		// wait for incoming client connections
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		// handle the connection
		go handleConnection(conn)
	}
}
