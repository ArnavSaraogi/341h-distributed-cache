package node

/*
Handles the networking stuff for the cache. Includes lru package, which handles the actual caching stuff.

Networking stuff:
	- Heartbeat to the configuration service (http endpoint?)
	- Initialize itself in the configuration service (http endpoint)
	- Communicate over TCP with client
*/

import (
	"distributedCache/lru"
	"log"
	"net"
	"strings"
)

type CacheNode struct {
	cache *lru.Cache
}

// initialize node
func NewNode(capacity int) *CacheNode {
	return &CacheNode{
		cache: lru.NewCache(capacity),
	}

	// TODO: INITIALIZE IN CONFIG SERVICE
}

// handle a connection to put or get
func (node *CacheNode) HandleConnection(conn net.Conn) {
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
			response := node.handleGet(parts)
			conn.Write([]byte(response))
		}
		if cmd == "PUT" {
			response := node.handlePut(parts)
			conn.Write([]byte(response))
		}
	}
}

// handle a PUT
func (node *CacheNode) handlePut(parts []string) string {
	if len(parts) < 3 {
		return "invalid request\n"
	}
	newkey := strings.TrimSpace(parts[1])
	newval := strings.TrimSpace(strings.Join(parts[2:], " "))
	node.cache.CachePut(newkey, newval)
	return "ok\n"
}

// handle a GET
func (node *CacheNode) handleGet(parts []string) string {
	key := strings.TrimSpace(parts[1])
	return node.cache.CacheGet(key)
}
