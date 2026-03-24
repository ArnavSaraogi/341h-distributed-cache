package node

/*
Handles the networking stuff for the cache. Includes lru package, which handles the actual caching stuff.

Networking stuff:
	- Heartbeat to the configuration service (http endpoint?)
	- Initialize itself in the configuration service (http endpoint)
	- Communicate over TCP with client
*/

import (
	"database/sql"
	"distributedCache/lru"
	"log"
	"net"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// logger
var logger = log.New(os.Stderr, "[CACHE SERVER]: ", log.Ltime)

type CacheNode struct {
	cache *lru.Cache
}

var base_url = os.Getenv("DATABASE_URL")

// initialize node
func NewNode(capacity int) *CacheNode {
	return &CacheNode{
		cache: lru.NewCache(capacity),
	}
}

// handle a connection to put or get
func (node *CacheNode) HandleConnection(conn net.Conn) {
	defer conn.Close()

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

	logger.Printf("Handled request %s", cmd)
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
	cand := node.cache.CacheGet(key)
	if cand == "" {
		db, err := sql.Open("postgres", os.Getenv(base_url))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		err = db.QueryRow("SELECT value FROM test_db WHERE key = $key", key).Scan(&key)
		if err != nil {
			log.Fatal(err)
		}
		cand = key
	}
	return cand
}
