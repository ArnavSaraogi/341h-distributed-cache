package main

import (
	"log"
	"net"
	"strings"
	"sync"
)

type SafeCache struct {
	mu    sync.Mutex
	cache map[string]string
}

func handleConnection(conn net.Conn, Cache *SafeCache) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}
		req := string(buf[:n])
		req = strings.TrimSpace(req)
		parts := strings.Fields(req)
		cmd := parts[0]
		if cmd == "GET" {
			response := handleGet(parts, Cache)
			conn.Write([]byte(response))
		}
		if cmd == "PUT" {
			response := handlePut(parts, Cache)
			conn.Write([]byte(response))
		}
	}
}

func handlePut(parts []string, Cache *SafeCache) string {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	newkey := strings.TrimSpace(parts[1])
	newval := strings.TrimSpace(parts[2])
	Cache.cache[newkey] = newval
	return "ok\n"
}

func handleGet(parts []string, Cache *SafeCache) string {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	key := strings.TrimSpace(parts[1])
	if val, ok := Cache.cache[key]; ok {
		return val
	} else {
		return "no such key"
	}
}

func main() {
	cache := SafeCache{
		cache: make(map[string]string),
	}
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
		go handleConnection(conn, &cache)
	}
}
